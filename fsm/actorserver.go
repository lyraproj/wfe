package fsm

import (
	"github.com/puppetlabs/go-fsm/api"
	"github.com/puppetlabs/go-issues/issue"
	"golang.org/x/net/context"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"reflect"
	"sort"
	"sync"
	"sync/atomic"
	"fmt"
	"github.com/puppetlabs/data-protobuf/datapb"
)

type ActorServer interface {
	api.Actor

	api.GoActorBuilder

	AddAction(na api.Action)

	// Validate ensures that all consumed values have a corresponding producer and that only
	// one producer exists for each produced value.
	Validate() error
}

type serverAction struct {
	api.Action
	graph.Node
	resolved chan bool
}

func (a *serverAction) SetResolved() {
	close(a.resolved)
}

func (a *serverAction) Resolved() <-chan bool {
	return a.resolved
}

type actorServer struct {
	context.Context
	name string
	input []api.Parameter
	output []api.Parameter
	actions map[string]api.Action

	runLatchLock sync.Mutex
	valuesLock   sync.RWMutex
	genesis      api.Genesis
	runLatch     map[int64]bool
	values       map[string]reflect.Value
	inbox        chan *serverAction
	jobCounter   int32
	done         chan bool
	graph        *simple.DirectedGraph
}

func NewActorServer(ctx context.Context, actorName string, input, output []api.Parameter) ActorServer {
	return &actorServer{
		Context:  ctx,
		name:     actorName,
		input:    input,
		output:   output,
		genesis:  NewGenesis(ctx),
		runLatch: make(map[int64]bool),
		values:   make(map[string]reflect.Value, 37),
		graph:    simple.NewDirectedGraph(),
		inbox:    make(chan *serverAction, 20),
		done:     make(chan bool)}
}

func (s *actorServer) Action(name string, function interface{}) {
	s.AddAction(api.NewGoAction(name, function))
}

func (s *actorServer) AddAction(na api.Action) {
	// Check that no other action is a producer of the same values
	for k, n := range s.graph.Nodes() {
		a := n.(api.Action)
		if a.Name() == na.Name() {
			panic(issue.NewReported(GENESIS_ACTION_ALREADY_DEFINED, issue.SEVERITY_ERROR, issue.H{`name`: na.Name()}, nil))
		}

		for _, ra := range a.Output() {
			for _, rb := range na.Output() {
				if ra.Name() == rb.Name() {
					panic(issue.NewReported(GENESIS_MULTIPLE_PRODUCERS_OF_VALUE, issue.SEVERITY_ERROR, issue.H{`name1`: k, `name2`: na.Name(), `value`: ra.Name}, nil))
				}
			}
		}
	}
	a := &serverAction{Action: na, Node: s.graph.NewNode(), resolved: make(chan bool)}
	s.graph.AddNode(a)
}

func (s *actorServer) Validate() error {
	// Build a map that associates a produced value with the producer of that value
	actions := s.graph.Nodes()
	valueProducers := make(map[string]api.Action, len(actions)*5)
	for _, a := range actions {
		fa := a.(api.Action)
		for _, r := range fa.Output() {
			valueProducers[r.Name()] = fa
		}
	}

	for _, a := range s.graph.Nodes() {
		fa := a.(api.Action)
		deps, err := s.dependents(fa, valueProducers)
		if err != nil {
			return err
		}
		for _, dep := range deps {
			s.graph.SetEdge(s.graph.NewEdge(dep.(graph.Node), a))
		}
	}
	return nil
}

func (s *actorServer) Input() []api.Parameter {
	return s.input
}

func (s *actorServer) Output() []api.Parameter {
	return s.output
}

func (s *actorServer) GetActions() map[string]api.Action {
	return s.actions
}

func (s *actorServer) Name() string {
	return s.name
}

func (s *actorServer) InvokeAction(actionName string, in map[string]reflect.Value, genesis api.Genesis) map[string]reflect.Value {
	action, found := s.actions[actionName]
	if !found {
		panic(fmt.Errorf("no action with name '%s' in actor '%s'", actionName, s.name))
	}
	return action.Call(genesis, in)
}

func (s *actorServer) Call(g api.Genesis, input map[string]reflect.Value) map[string]reflect.Value {
	// Run all nodes that can run, i.e. root nodes
	actions := s.graph.Nodes()
	if len(actions) == 0 {
		return nil
	}

	params := s.input
	args := make(map[string]reflect.Value, len(params))
	s.valuesLock.RLock()
	for _, param := range params {
		n := param.Name()
		lu := param.Lookup()
		var v reflect.Value
		ok := false
		if lu == nil {
			v, ok = input[n]
		} else {
			v, ok = s.lookupOne(lu.String())
		}
		if !ok {
			panic(issue.NewReported(GENESIS_NO_PRODUCER_OF_VALUE, issue.SEVERITY_ERROR, issue.H{`action`: s.name, `value`: n}, nil))
		}
		args[n] = v
	}
	s.valuesLock.RUnlock()

	for w := 1; w <= 5; w++ {
		go s.actionWorker(w)
	}
	for _, n := range actions {
		if len(s.graph.To(n.ID())) == 0 {
			s.scheduleAction(n.(*serverAction))
		}
	}
	<-s.done

	result := make(map[string]reflect.Value, len(s.output))
	for _, out := range s.output {
		n := out.Name()
		result[n] = s.values[n]
	}
	return result
}

func (s *actorServer) DumpVariables() {
	names := make([]string, 0, len(s.values))
	for n := range s.values {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		fmt.Printf("%s = %v\n", n, s.values[n])
	}
}

func (s *actorServer) dependents(a api.Action, valueProducers map[string]api.Action) ([]api.Action, error) {
	dam := make(map[string]api.Action, 0)
	for _, param := range a.Input() {
		if param.Lookup() == nil {
			if vp, found := valueProducers[param.Name()]; found {
				dam[vp.Name()] = vp
				continue
			}
			return nil, issue.NewReported(GENESIS_NO_PRODUCER_OF_VALUE, issue.SEVERITY_ERROR, issue.H{`action`: a.Name(), `value`: param.Name}, nil)
		}
	}
	da := make([]api.Action, 0, len(dam))
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
func (s *actorServer) actionWorker(id int) {
	for a := range s.inbox {
		s.runAction(a)
		if atomic.AddInt32(&s.jobCounter, -1) <= 0 {
			close(s.inbox)
			close(s.done)
		}
	}
}

func (s *actorServer) runAction(a *serverAction) {
	s.runLatchLock.Lock()
	if s.runLatch[a.ID()] {
		return
	}
	s.runLatch[a.ID()] = true
	s.runLatchLock.Unlock()

	s.waitForEdgesTo(a)

	params := a.Input()
	args := make(map[string]reflect.Value, len(params))
	s.valuesLock.RLock()
	for _, param := range params {
		n := param.Name()
		lu := param.Lookup()
		if lu == nil {
			args[n] = s.values[n]
		} else {
			if v, ok := s.lookupOne(lu.String()); ok {
				args[n] = v
			} else {
				panic(issue.NewReported(GENESIS_NO_PRODUCER_OF_VALUE, issue.SEVERITY_ERROR, issue.H{`action`: a.Name(), `value`: n}, nil))
			}
		}
	}
	s.valuesLock.RUnlock()

	result := a.Call(s.genesis, args)
	if result != nil && len(result) > 0 {
		s.valuesLock.Lock()
		for k, v := range result {
			s.values[k] = v
		}
		s.valuesLock.Unlock()
	}
	a.SetResolved()

	// Schedule all actions that are dependent on this action. Since a node can be
	// dependent on several actions, it might be scheduled several times. It will
	// however only run once. This is controlled by the runLatch.
	for _, n := range s.graph.From(a.ID()) {
		s.scheduleAction(n.(*serverAction))
	}
}

func (s *actorServer) Lookup(keys []string) map[string]reflect.Value {
	result := make(map[string]reflect.Value, len(keys))
	for _, k := range keys {
		if v, ok := s.lookupOne(k); ok {
			result[k] = v
		}
	}
	return result
}

func (s *actorServer) lookupOne(key string) (reflect.Value, bool) {
	testing := map[string]interface{}{
		`aws.region`: `eu-west-1`,
		`aws.tags`: map[string]string {
			`created_by`: `john.mccabe@puppet.com`,
  		`department`: `engineering`,
			`project`   : `incubator`,
	  	`lifetime`  : `1h`,
		},
	}
	v, ok := testing[key]
	if ok {
		return reflect.ValueOf(v), true
	}
	return datapb.InvalidValue, false
}

// Ensure that all nodes that has an edge to this node have been
// fully resolved.
func (s *actorServer) waitForEdgesTo(a *serverAction) {
	parents := s.graph.To(a.ID())
	for _, before := range parents {
		<-before.(*serverAction).Resolved()
	}
}

func (s *actorServer) scheduleAction(a *serverAction) {
	atomic.AddInt32(&s.jobCounter, 1)
	s.inbox <- a
}
