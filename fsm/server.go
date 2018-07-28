package fsm

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-issues/issue"
	"gonum.org/v1/gonum/graph/simple"
	"reflect"
	"sort"
	"sync"
	"sync/atomic"
	"gonum.org/v1/gonum/graph"
)

type Genesis interface {
	Action(name string, function interface{}, paramNames ...string)
}

type GenesisService interface {
	eval.Context
	Genesis
	AddAction(na Action)
	Validate() error
	Run() error
	RunAction(ServerAction)
}

type genesisImpl struct {
	eval.Context
	runLatchLock sync.Mutex
	valuesLock   sync.RWMutex
	runLatch     map[int64]bool
	values       map[string]reflect.Value
	inbox        chan ServerAction
	jobCounter   int32
	done         chan bool
	graph        *simple.DirectedGraph
}

func GetGenesisService(ctx eval.Context) GenesisService {
	return &genesisImpl{Context: ctx, runLatch: make(map[int64]bool), values: make(map[string]reflect.Value, 37), graph: simple.NewDirectedGraph(), done: make(chan bool)}
}

func (g *genesisImpl) Action(name string, function interface{}, paramNames ...string) {
	g.AddAction(NewGoAction(g, name, function, paramNames))
}

func (g *genesisImpl) AddAction(na Action) {
	// Check that no other action is a producer of the same values
	for k, n := range g.graph.Nodes() {
		a := n.(Action)
		if a.Name() == na.Name() {
			panic(eval.Error(g, GENESIS_ACTION_ALREADY_DEFINED, issue.H{`name`: na.Name()}))
		}

		for _, ra := range a.Produces() {
			for _, rb := range na.Produces() {
				if ra.Name() == rb.Name() {
					panic(eval.Error(g, GENESIS_MULTIPLE_PRODUCERS_OF_VALUE, issue.H{`name1`: k, `name2`: na.Name(), `value`: ra.Name()}))
				}
			}
		}
	}
	a := &serverAction{Action: na, Node: g.graph.NewNode(), resolved: make(chan bool)}
	g.graph.AddNode(a)
}

func (g *genesisImpl) Validate() error {
	// Build a map that associates a produced value with the producer of that value
	actions := g.graph.Nodes()
	valueProducers := make(map[string]Action, len(actions)*5)
	for _, a := range actions {
		for _, r := range a.(Action).Produces() {
			valueProducers[r.Name()] = a.(Action)
		}
	}

	for _, a := range g.graph.Nodes() {
		deps, err := g.dependents(a.(Action), valueProducers)
		if err != nil {
			return err
		}
		for _, dep := range deps {
			g.graph.SetEdge(g.graph.NewEdge(dep.(graph.Node), a))
		}
	}
	return nil
}

func (g *genesisImpl) dependents(a Action, valueProducers map[string]Action) ([]Action, error) {
	dam := make(map[string]Action, 0)
	for _, param := range a.Consumes() {
		if vp, found := valueProducers[param.Name()]; found {
			dam[vp.Name()] = vp
			continue
		}
		return nil, eval.Error(g, GENESIS_NO_PRODUCER_OF_VALUE, issue.H{`action`: a.Name(), `value`: param.name})
	}
	da := make([]Action, 0, len(dam))
	for _, vp := range dam {
		da = append(da, vp)
	}

	// Ensure that actions are sorted by name
	sort.Slice(da, func(i, j int) bool {
		return da[i].Name() < da[j].Name()
	})
	return da, nil
}

func (g *genesisImpl) Run() error {
	// Run all nodes that can run, i.e. root nodes
	g.initialize()
	actions := g.graph.Nodes()
	if len(actions) == 0 {
		return nil
	}
	for _, n := range actions {
		if len(g.graph.To(n.ID())) == 0 {
			g.scheduleAction(n.(ServerAction))
		}
	}
	<-g.done
	return nil
}

func (g *genesisImpl) initialize() {
	g.inbox = make(chan ServerAction, 20)
	for w := 1; w <= 5; w++ {
		go g.actionWorker(w)
	}
}

// This function represents a worker that spawns actions
func (g *genesisImpl) actionWorker(id int) {
	for a := range g.inbox {
		g.RunAction(a)
		if atomic.AddInt32(&g.jobCounter, -1) <= 0 {
			close(g.inbox)
			close(g.done)
		}
	}
}

func (g *genesisImpl) RunAction(a ServerAction) {
	g.runLatchLock.Lock()
	if g.runLatch[a.ID()] {
		return
	}
	g.runLatch[a.ID()] = true
	g.runLatchLock.Unlock()

	g.waitForEdgesTo(a)

	params := a.Consumes()
	args := make([]reflect.Value, len(params))
	g.valuesLock.RLock()
	for i, param := range params {
		args[i] = g.values[param.name]
	}
	g.valuesLock.RUnlock()

	result, err := a.Call(g, args)
	if err != nil {
		panic(err)
	}
	if result != nil {
		g.valuesLock.Lock()
		for k, v := range result {
			g.values[k] = v
		}
		g.valuesLock.Unlock()
	}
	a.SetResolved()

	// Schedule all actions that are dependent on this action. Since a node can be
	// dependent on several actions, it might be scheduled several times. It will
	// however only run once. This is controlled by the runLatch.
	for _, n := range g.graph.From(a.ID()) {
		g.scheduleAction(n.(ServerAction))
	}
}

// Ensure that all nodes that has an edge to this node have been
// fully resolved.
func (g *genesisImpl) waitForEdgesTo(a ServerAction) {
	graph := g.graph
	parents := graph.To(a.ID())
	for _, before := range parents {
		<-before.(ServerAction).Resolved()
	}
}

func (g* genesisImpl) scheduleAction(a ServerAction) {
	atomic.AddInt32(&g.jobCounter, 1)
	g.inbox <- a
}