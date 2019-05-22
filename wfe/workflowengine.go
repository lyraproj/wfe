package wfe

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/hashicorp/go-hclog"

	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/types"
	"github.com/lyraproj/servicesdk/serviceapi"
	"github.com/lyraproj/servicesdk/wf"
	"github.com/lyraproj/wfe/api"
	"github.com/lyraproj/wfe/service"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/encoding"
	"gonum.org/v1/gonum/graph/encoding/dot"
	"gonum.org/v1/gonum/graph/simple"
)

type WorkflowEngine interface {
	Run(ctx px.Context, parameters px.OrderedMap) px.OrderedMap

	BuildInvertedGraph(ctx px.Context, existsFunc func(string) bool)

	GraphAsDot() []byte

	// Validate ensures that all consumed values have a corresponding producer and that only
	// one producer exists for each produced value.
	Validate()
}

type serverStep struct {
	api.Step
	graph.Node
	resolved chan bool
}

func appendParameterNames(params []serviceapi.Parameter, b *bytes.Buffer) {
	for i, p := range params {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(p.Name())
	}
}

func (a *serverStep) Attributes() []encoding.Attribute {
	b := bytes.NewBufferString(`"`)
	b.WriteString(service.LeafName(a.Name()))
	b.WriteByte('{')
	b.WriteString("\nparameters:[")
	appendParameterNames(a.Parameters(), b)
	b.WriteString("],\nreturns:[")
	appendParameterNames(a.Returns(), b)
	b.WriteString(`]}"`)
	return []encoding.Attribute{{Key: "label", Value: b.String()}}
}

func (a *serverStep) DOTID() string {
	return service.LeafName(a.Name())
}

func (a *serverStep) SetResolved() {
	close(a.resolved)
}

func (a *serverStep) Resolved() <-chan bool {
	return a.resolved
}

type workflowEngine struct {
	api.Workflow
	runLatchLock sync.Mutex
	valuesLock   sync.RWMutex
	runLatch     map[int64]bool
	values       map[string]px.Value
	inbox        chan *serverStep
	jobCounter   int32
	done         chan bool
	graph        *simple.DirectedGraph
	errors       []error
}

func NewWorkflowEngine(workflow api.Workflow) WorkflowEngine {
	as := &workflowEngine{
		Workflow: workflow,
		runLatch: make(map[int64]bool),
		graph:    simple.NewDirectedGraph(),
		inbox:    make(chan *serverStep, 20),
		done:     make(chan bool)}

	for _, a := range workflow.Steps() {
		as.addStep(a)
	}
	return as
}

func (s *workflowEngine) addStep(na api.Step) {
	// Check that no other step is a producer of the same values
	ni := s.graph.Nodes()
	if ni != nil {
		for ni.Next() {
			a := ni.Node().(api.Step)
			if a.Name() == na.Name() {
				panic(px.Error(AlreadyDefined, issue.H{`name`: na.Name()}))
			}
		}
	}
	a := &serverStep{Step: na, Node: s.graph.NewNode(), resolved: make(chan bool)}
	s.graph.AddNode(a)
}

// maxGuards control how many possible variations there can be of the workflow graph. The
// actual number is 2 to the power maxGuards.
const maxGuards = 8

func (s *workflowEngine) GraphAsDot() []byte {
	de, err := dot.Marshal(s.graph, s.Name(), ``, `  `)
	if err != nil {
		panic(px.Error(GraphDotMarshal, issue.H{`detail`: err.Error()}))
	}
	return de
}

func (s *workflowEngine) BuildInvertedGraph(c px.Context, existsFunc func(string) bool) {
	g := s.graph
	ni := g.Nodes()
	if ni == nil {
		return
	}

	ei := g.Edges()
	for ei.Next() {
		e := ei.Edge()
		g.RemoveEdge(e.From().ID(), e.To().ID())
	}

	// Add workflow as the producer of parameters with values.
	vp := make(valueProducers, ni.Len()*5)
	vp.add(s, s.Parameters())
	for ni.Next() {
		fa := ni.Node().(*serverStep)
		if fa.When() == wf.Always || existsFunc(fa.Identifier()) {
			vp.add(fa, fa.Returns())
		}
	}

	ni.Reset()
	for ni.Next() {
		fa := ni.Node().(*serverStep)
		if fa.When() == wf.Always || existsFunc(fa.Identifier()) {
			ds := s.dependents(fa, vp)
			for _, dep := range ds {
				g.SetEdge(g.NewEdge(fa, dep.(graph.Node)))
			}
		}
	}
}

func (s *workflowEngine) Validate() {
	// Build a map that associates a produced value with the producer of that value
	guards := make(map[string]bool)

	ni := s.graph.Nodes()
	if ni == nil {
		return
	}

	for ni.Next() {
		for _, g := range ni.Node().(*serverStep).When().Names() {
			guards[g] = false
		}
	}

	gc := uint(len(guards))
	if gc > 0 {
		maxVariations := int(math.Pow(2.0, float64(gc)))
		if gc > maxGuards {
			panic(px.Error(TooManyGuards, issue.H{`step`: s, `max`: maxGuards, `count`: gc}))
		}

		guardNames := make([]string, 0, gc)
		for n := range guards {
			guardNames = append(guardNames, n)
		}
		sort.Strings(guardNames)

		// Check all variations for validity with respect to parameters and returns
		for bitmap := 0; bitmap <= maxVariations; bitmap++ {
			es := make([]*types.HashEntry, gc)
			for i := uint(0); i < gc; i++ {
				es[i] = types.WrapHashEntry2(guardNames[i], types.WrapBoolean(bitmap&(1<<i) == 1))
			}
			guardCombo := types.WrapHash(es)

			// Add workflow as the producer of parameters with values.
			ni.Reset()
			vp := make(valueProducers, ni.Len()*5)
			vp.add(s, s.Parameters())

			for ni.Next() {
				fa := ni.Node().(*serverStep)
				if fa.When().IsTrue(guardCombo) {
					vp.add(fa, fa.Returns())
				}
			}

			ni.Reset()
			for ni.Next() {
				vp.validateParameters(ni.Node().(*serverStep))
			}
			vp.validate(s)
		}

		// Build the final graph that doesn't care about guards or multiple producers of a value
		ni.Reset()
		vp := make(valueProducers, ni.Len()*5)
		vp.add(s, s.Parameters())
		for ni.Next() {
			fa := ni.Node().(*serverStep)
			vp.add(fa, fa.Returns())
		}

		ni.Reset()
		for ni.Next() {
			fa := ni.Node().(*serverStep)
			ds := s.dependents(fa, vp)
			for _, dep := range ds {
				s.graph.SetEdge(s.graph.NewEdge(dep.(graph.Node), fa))
			}
		}
	} else {
		// Add workflow as the producer of parameters with values.
		ni.Reset()
		vp := make(valueProducers, ni.Len()*5)
		vp.add(s, s.Parameters())
		for ni.Next() {
			fa := ni.Node().(*serverStep)
			vp.add(fa, fa.Returns())
		}

		ni.Reset()
		for ni.Next() {
			fa := ni.Node().(*serverStep)
			ds := s.dependents(fa, vp)
			for _, dep := range ds {
				s.graph.SetEdge(s.graph.NewEdge(dep.(graph.Node), fa))
			}
		}
		vp.validate(s)
	}
}

type valueProducers map[string][]api.Step

func (vp valueProducers) add(a api.Step, ps []serviceapi.Parameter) {
	for _, param := range ps {
		n := param.Name()
		v := vp[n]
		if v == nil {
			vp[n] = []api.Step{a}
		} else {
			vp[n] = append(v, a)
		}
	}
}

func (vp valueProducers) validate(a api.Step) {
	for k, v := range vp {
		if len(v) > 1 {
			panic(px.Error(MultipleProducersOfValue, issue.H{`step1`: v[0], `step2`: v[1], `value`: k}))
		}
	}
	for _, param := range a.Returns() {
		if _, found := vp[param.Name()]; found {
			continue
		}
		panic(px.Error(NoProducerOfValue, issue.H{`step`: a, `value`: param.Name()}))
	}
}

func (vp valueProducers) validateParameters(a api.Step) {
	var checkDep = func(name string) {
		if _, found := vp[name]; !found {
			panic(px.Error(NoProducerOfValue, issue.H{`step`: a, `value`: name}))
		}
	}
	for _, name := range a.When().Names() {
		checkDep(name)
	}
	for _, param := range a.Parameters() {
		if param.Value() == nil {
			checkDep(param.Name())
		}
	}
}

func (s *workflowEngine) Run(ctx px.Context, parameters px.OrderedMap) px.OrderedMap {
	s.values = make(map[string]px.Value, 37)
	ResolveParameters(ctx, s.Workflow, parameters).EachPair(func(k, v px.Value) {
		s.values[k.String()] = v
	})

	// Run all nodes that can run, i.e. root nodes
	ni := s.graph.Nodes()
	if ni == nil || ni.Len() == 0 {
		return nil
	}

	for w := 1; w <= 5; w++ {
		px.Fork(ctx, func(cf px.Context) { s.stepWorker(cf, w) })
	}
	for ni.Next() {
		n := ni.Node()
		if s.graph.To(n.ID()).Len() == 0 {
			s.scheduleStep(n.(*serverStep))
		}
	}
	<-s.done

	if s.errors != nil {
		var err error
		if len(s.errors) == 1 {
			err = s.errors[0]
		} else {
			err = px.Error(api.MultipleErrors, issue.H{`errors`: s.errors})
		}
		panic(err)
	}

	entries := make([]*types.HashEntry, 0, len(s.Returns()))
	for _, out := range s.Returns() {
		n := out.Name()
		if v, ok := s.values[n]; ok {
			entries = append(entries, types.WrapHashEntry2(n, v))
		}
	}
	return types.WrapHash(entries)
}

func (s *workflowEngine) DumpVariables() {
	names := make([]string, 0, len(s.values))
	for n := range s.values {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		fmt.Printf("%s = %v\n", n, s.values[n])
	}
}

func (s *workflowEngine) dependents(a api.Step, vp valueProducers) []api.Step {

	dam := make(map[string]api.Step)
	var addDeps = func(name string) {
		if ds, found := vp[name]; found {
			for _, d := range ds {
				if d != s { // Workflow itself only has external dependencies
					dam[d.Name()] = d
				}
			}
			return
		}
		panic(px.Error(NoProducerOfValue, issue.H{`step`: a, `value`: name}))
	}

nextName:
	for _, name := range a.When().Names() {
		for _, param := range a.Parameters() {
			if name == param.Name() {
				continue nextName
			}
		}
		addDeps(name)
	}
	for _, param := range a.Parameters() {
		if param.Value() == nil {
			addDeps(param.Name())
		}
	}

	da := make([]api.Step, 0, len(dam))
	for _, vp := range dam {
		da = append(da, vp)
	}

	// Ensure that steps are sorted by name
	sort.Slice(da, func(i, j int) bool {
		return da[i].Name() < da[j].Name()
	})
	return da
}

// This function represents a worker that spawns steps
func (s *workflowEngine) stepWorker(ctx px.Context, id int) {
	for a := range s.inbox {
		s.runStep(ctx, a)
	}
}

// This value should never be made visible. It's only purpose is to prevent
// execution of steps that detects it on input
var unresolvedValue = types.WrapString(`unresolved_` + strconv.Itoa(rand.Int()))

func (s *workflowEngine) runStep(ctx px.Context, a *serverStep) {
	defer func() {
		if atomic.AddInt32(&s.jobCounter, -1) <= 0 {
			close(s.inbox)
			close(s.done)
		}
	}()

	s.runLatchLock.Lock()
	if s.runLatch[a.ID()] {
		s.runLatchLock.Unlock()
		return
	}
	s.runLatch[a.ID()] = true
	s.runLatchLock.Unlock()

	result := s.runStepCatch(ctx, a)

	s.valuesLock.Lock()
	for k, v := range result {
		s.values[k] = v
	}
	s.valuesLock.Unlock()

	a.SetResolved()

	// Schedule all steps that are dependent on this step. Since a node can be
	// dependent on several steps, it might be scheduled several times. It will
	// however only run once. This is controlled by the runLatch.
	ni := s.graph.From(a.ID())
	for ni.Next() {
		s.scheduleStep(ni.Node().(*serverStep))
	}
}

var hasLocationPattern = regexp.MustCompile(`[^\w]file:\s+|[^\w]line:\s+`)

type parameterScope struct {
	parent px.Keyed
	params map[string]px.Value
}

func (ps *parameterScope) Get(key px.Value) (px.Value, bool) {
	if v, ok := ps.params[key.String()]; ok {
		return v, true
	}
	return ps.parent.Get(key)
}

func (s *workflowEngine) runStepCatch(ctx px.Context, a *serverStep) (returns map[string]px.Value) {
	defer func() {
		r := recover()
		if r != nil {
			var err error
			var loc issue.Location
			if orig := s.Origin(); orig != `` {
				loc = issue.ParseLocation(orig)
			}
			switch r := r.(type) {
			case issue.Reported:
				if loc != nil {
					r = r.WithLocation(loc)
				}
				err = r
			case error:
				if loc := s.Origin(); loc != `` {
					es := r.Error()
					if !hasLocationPattern.MatchString(es) {
						r = fmt.Errorf(`%s %s`, es, loc)
					}
				}
				err = r
			case string:
				if loc := s.Origin(); loc != `` {
					err = fmt.Errorf(`%s %s`, r, loc)
				} else {
					err = errors.New(r)
				}
			case fmt.Stringer:
				if loc := s.Origin(); loc != `` {
					err = fmt.Errorf(`%s %s`, r.String(), loc)
				} else {
					err = errors.New(r.String())
				}
			default:
				err = fmt.Errorf("%v", r)
			}
			if _, ok := err.(issue.Reported); !ok {
				// Wrap in a step execution error so that step is revealed
				err = issue.NewNested(StepExecutionError, issue.H{`step`: a}, loc, err)
			}
			s.runLatchLock.Lock()
			if s.errors == nil {
				s.errors = []error{err}
			} else {
				s.errors = append(s.errors, err)
			}
			s.runLatchLock.Unlock()
			for _, p := range a.Returns() {
				returns[p.Name()] = unresolvedValue
			}
		}
	}()

	returns = make(map[string]px.Value, len(a.Returns()))

	s.waitForEdgesTo(a)

	params := a.Parameters()
	scope := &parameterScope{ctx.Scope(), make(map[string]px.Value, len(params))}
	entries := make([]*types.HashEntry, len(params))
	for i, param := range params {
		pv := s.resolveParameter(ctx, a, param, scope)
		if pv == unresolvedValue {
			// This step cannot run so the returned values are all unresolved
			hclog.Default().Info(`skipping step because of earlier errors`, `name`, a.Label())
			for _, p := range a.Returns() {
				returns[p.Name()] = unresolvedValue
			}
			return
		}
		scope.params[param.Name()] = pv
		entries[i] = types.WrapHashEntry2(param.Name(), pv)
	}

	// Run the step
	r := a.Run(ctx, types.WrapHash(entries)).(px.OrderedMap)
	if r == nil {
		r = px.EmptyMap
	}
	for _, p := range a.Returns() {
		n := p.Name()
		if v, ok := r.Get4(n); ok {
			returns[n] = v
		} else {
			panic(px.Error(ExpectedValueNotProduced, issue.H{`step`: a, `value`: n}))
		}
	}
	return
}

func (s *workflowEngine) resolveParameter(ctx px.Context, step api.Step, param serviceapi.Parameter, scope px.Keyed) px.Value {
	n := param.Name()
	if param.Value() == nil {
		s.valuesLock.RLock()
		v, ok := s.values[n]
		s.valuesLock.RUnlock()
		if ok {
			return v
		}
		panic(px.Error(NoProducerOfValue, issue.H{`step`: step, `value`: n}))
	}
	return types.ResolveDeferred(ctx, param.Value(), scope)
}

// Ensure that all nodes that has an edge to this node have been
// fully resolved.
func (s *workflowEngine) waitForEdgesTo(a *serverStep) {
	parents := s.graph.To(a.ID())
	for parents.Next() {
		<-parents.Node().(*serverStep).Resolved()
	}
}

func (s *workflowEngine) scheduleStep(a *serverStep) {
	atomic.AddInt32(&s.jobCounter, 1)
	s.inbox <- a
}
