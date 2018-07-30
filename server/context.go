package server

import (
	"github.com/puppetlabs/go-fsm/fsm"
	"github.com/puppetlabs/go-issues/issue"
	"golang.org/x/net/context"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"reflect"
	"sort"
	"sync"
	"sync/atomic"
)

type Context interface {
	fsm.Context

	AddAction(na fsm.Action)

	// Validate ensures that all consumed values have a corresponding producer and that only
	// one procuder exists for each produced value.
	Validate() error

	// Run runs all registered actions in the order determined by their produces/consumes
	Run() error
}

type fsmContext struct {
	context.Context
	runLatchLock sync.Mutex
	valuesLock   sync.RWMutex
	runLatch     map[int64]bool
	values       map[string]reflect.Value
	inbox        chan ServerAction
	jobCounter   int32
	done         chan bool
	graph        *simple.DirectedGraph
}

func NewContext(ctx context.Context) Context {
	return &fsmContext{
		Context:  ctx,
		runLatch: make(map[int64]bool),
		values:   make(map[string]reflect.Value, 37),
		graph:    simple.NewDirectedGraph(),
		inbox:    make(chan ServerAction, 20),
		done:     make(chan bool)}
}

func (g *fsmContext) Action(name string, function interface{}) {
	g.AddAction(fsm.NewGoAction(name, function))
}

func (g *fsmContext) AddAction(na fsm.Action) {
	// Check that no other action is a producer of the same values
	for k, n := range g.graph.Nodes() {
		a := n.(fsm.Action)
		if a.Name() == na.Name() {
			panic(issue.NewReported(fsm.GENESIS_ACTION_ALREADY_DEFINED, issue.SEVERITY_ERROR, issue.H{`name`: na.Name()}, nil))
		}

		for _, ra := range a.Produces() {
			for _, rb := range na.Produces() {
				if ra.Name == rb.Name {
					panic(issue.NewReported(fsm.GENESIS_MULTIPLE_PRODUCERS_OF_VALUE, issue.SEVERITY_ERROR, issue.H{`name1`: k, `name2`: na.Name(), `value`: ra.Name}, nil))
				}
			}
		}
	}
	a := &serverAction{Action: na, Node: g.graph.NewNode(), resolved: make(chan bool)}
	g.graph.AddNode(a)
}

func (g *fsmContext) Validate() error {
	// Build a map that associates a produced value with the producer of that value
	actions := g.graph.Nodes()
	valueProducers := make(map[string]fsm.Action, len(actions)*5)
	for _, a := range actions {
		fa := a.(fsm.Action)
		for _, r := range fa.Produces() {
			valueProducers[r.Name] = fa
		}
	}

	for _, a := range g.graph.Nodes() {
		fa := a.(fsm.Action)
		deps, err := g.dependents(fa, valueProducers)
		if err != nil {
			return err
		}
		for _, dep := range deps {
			g.graph.SetEdge(g.graph.NewEdge(dep.(graph.Node), a))
		}
	}
	return nil
}

func (g *fsmContext) Run() error {
	// Run all nodes that can run, i.e. root nodes
	actions := g.graph.Nodes()
	if len(actions) == 0 {
		return nil
	}

	for w := 1; w <= 5; w++ {
		go g.actionWorker(w)
	}
	for _, n := range actions {
		if len(g.graph.To(n.ID())) == 0 {
			g.scheduleAction(n.(ServerAction))
		}
	}
	<-g.done
	return nil
}

func (g *fsmContext) dependents(a fsm.Action, valueProducers map[string]fsm.Action) ([]fsm.Action, error) {
	dam := make(map[string]fsm.Action, 0)
	for _, param := range a.Consumes() {
		if vp, found := valueProducers[param.Name]; found {
			dam[vp.Name()] = vp
			continue
		}
		return nil, issue.NewReported(fsm.GENESIS_NO_PRODUCER_OF_VALUE, issue.SEVERITY_ERROR, issue.H{`action`: a.Name(), `value`: param.Name}, nil)
	}
	da := make([]fsm.Action, 0, len(dam))
	for _, vp := range dam {
		da = append(da, vp)
	}

	// Ensure that actions are sorted by name
	sort.Slice(da, func(i, j int) bool {
		return da[i].Name() < da[j].Name()
	})
	return da, nil
}

// This function represents a worker that spawns actions
func (g *fsmContext) actionWorker(id int) {
	for a := range g.inbox {
		g.runAction(a)
		if atomic.AddInt32(&g.jobCounter, -1) <= 0 {
			close(g.inbox)
			close(g.done)
		}
	}
}

func (g *fsmContext) runAction(a ServerAction) {
	g.runLatchLock.Lock()
	if g.runLatch[a.ID()] {
		return
	}
	g.runLatch[a.ID()] = true
	g.runLatchLock.Unlock()

	g.waitForEdgesTo(a)

	params := a.Consumes()
	args := make(map[string]reflect.Value, len(params))
	g.valuesLock.RLock()
	for _, param := range params {
		args[param.Name] = g.values[param.Name]
	}
	g.valuesLock.RUnlock()

	result := a.Call(g, args)
	if result != nil && len(result) > 0 {
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
func (g *fsmContext) waitForEdgesTo(a ServerAction) {
	parents := g.graph.To(a.ID())
	for _, before := range parents {
		<-before.(ServerAction).Resolved()
	}
}

func (g *fsmContext) scheduleAction(a ServerAction) {
	atomic.AddInt32(&g.jobCounter, 1)
	g.inbox <- a
}
