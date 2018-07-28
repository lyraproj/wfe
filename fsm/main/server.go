package main

import (
	"net"
	"log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"golang.org/x/net/context"
	"github.com/puppetlabs/go-fsm/fsm/fsmpb"
	"github.com/puppetlabs/data-protobuf/datapb"
	"github.com/puppetlabs/go-evaluator/eval"
	"reflect"
	"github.com/puppetlabs/go-issues/issue"
	"github.com/puppetlabs/go-fsm/fsm"
	"fmt"
	"github.com/puppetlabs/go-fsm/misc"

	// Initialize pcore
	_ "github.com/puppetlabs/go-evaluator/pcore"
)

type server struct{
	eval.Context
	actions []fsm.Action
}

func (s *server) GetActor(ctx context.Context, request *fsmpb.ActorRequest) (*fsmpb.ActorResponse, error) {
	aa := make([]*fsmpb.Action, len(s.actions))
	for i, a := range s.actions {
		aa[i] = &fsmpb.Action{Id: int64(i), Name: a.Name(), Consumes: s.convertParams(a.Consumes()), Produces: s.convertParams(a.Produces())}
	}
	return &fsmpb.ActorResponse{Actions: aa}, nil
}

func (s *server) InvokeAction(ctx context.Context, in *fsmpb.ActionInvocation) (*datapb.DataHash, error) {
	id := int(in.Id)
	if id < 0 || id >= len(s.actions) {
		return nil, fmt.Errorf("no action with ID %d", id)
	}

	// Convert named args to positional
	a := s.actions[id]
	hash := misc.FromPBData(&datapb.Data{Kind: &datapb.Data_HashValue{in.Arguments}}).(eval.KeyedValue)
	rm := make([]reflect.Value, len(a.Consumes()))
	rf := s.Reflector()
	for i, p := range a.Consumes() {
		if v, ok := hash.Get4(p.Name()); ok {
			if vt, ok := rf.ReflectType(v.Type()); ok {
				rm[i] = rf.Reflect(v, vt)
			} else {
				panic(eval.Error(s, fsm.GENESIS_UNABLE_TO_REFLECT_TYPE, issue.H{`type`: vt}))
			}
		}
	}
	result, err := a.Call(s, rm)
	if err != nil {
		return nil, err
	}
	return misc.ToPBData(eval.Wrap2(s, result)).GetHashValue(), nil
}

func (s *server) Action(name string, function interface{}, paramNames ...string) {
	s.actions = append(s.actions, fsm.NewGoAction(s, name, function, paramNames))
}

func (s *server) convertParams(parameters []*fsm.Parameter) []*fsmpb.Parameter {
	fp := make([]*fsmpb.Parameter, len(parameters))
	for i, p := range parameters {
		fp[i] = &fsmpb.Parameter{p.Name(), p.Type().String()}
	}
	return fp
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	genesis := &server{Context: eval.Puppet.RootContext(), actions: make([]fsm.Action, 0)}
	fsmpb.RegisterActorServer(s, genesis)
	// Register reflection service on gRPC server.
	reflection.Register(s)

	TestTheThing(genesis)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

type StartActionResult struct {
	A string
	B int64
}

type ActionAResult struct {
	C string
	D int64
}

type ActionBResult struct {
	E string
	F int64
}

func TestTheThing(genesis fsm.Genesis) {
	genesis.Action("a", func(g fsm.Genesis) (*StartActionResult, error) {
		return &StartActionResult{`hello`, 4}, nil
	})

	genesis.Action("b1", func(g fsm.Genesis, a string, b int64) (*ActionAResult, error) {
		return &ActionAResult{a + ` world`, b + 4}, nil
	}, `a.a`, `a.b`)

	genesis.Action("b2", func(g fsm.Genesis, a string, b int64) (*ActionBResult, error) {
		return &ActionBResult{a + ` earth`, b + 8}, nil
	}, `a.a`, `a.b`)

	genesis.Action("c", func(g fsm.Genesis, c string, d int64, e string, f int64) error {
		fmt.Printf("%s, %d, %s, %d\n", c, d, e, f)
		return nil
	}, `b1.c`, `b1.d`, `b2.e`, `b2.f`)
}