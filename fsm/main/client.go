package main

import (
	"github.com/puppetlabs/data-protobuf/datapb"
	"github.com/puppetlabs/go-fsm/fsm"
	"github.com/puppetlabs/go-fsm/fsm/fsmpb"
	"github.com/puppetlabs/go-fsm/misc"
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/types"
	"github.com/puppetlabs/go-issues/issue"
	"google.golang.org/grpc"
	"log"
	"reflect"

	// Initialize pcore
	_ "github.com/puppetlabs/go-evaluator/pcore"
)

const (
	address = "localhost:50051"
)

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	eval.Puppet.Do(func(ctx eval.Context) error {
		client := fsmpb.NewActorClient(conn)
		actions, err := client.GetActor(ctx, &fsmpb.ActorRequest{Name: `default`})
		if err != nil {
			return err
		}

		g := fsm.GetGenesisService(ctx)
		for _, action := range actions.Actions {
			g.AddAction(fsm.NewAction(action.Name, &grpcActionCall{client, action.Id}, convertParams(g, action.Consumes), convertParams(g, action.Produces)))
		}
		g.Validate()
		g.Run()
		return nil
	})
}

type grpcActionCall struct {
	client fsmpb.ActorClient
	id     int64
}

func (ga *grpcActionCall) Call(g fsm.Genesis, a fsm.Action, args []reflect.Value) (map[string]reflect.Value, error) {
	nargs := len(args)
	c := g.(eval.Context)
	if nargs != len(a.Consumes()) {
		panic(eval.Error(c, fsm.GENESIS_ACTION_BAD_CONSUMES_COUNT, issue.H{`name`: a.Name(), `expected`: len(a.Consumes()), `actual`: nargs}))
	}

	entries := make([]*types.HashEntry, nargs)
	for i, p := range a.Consumes() {
		entries[i] = types.WrapHashEntry2(p.Name(), eval.Wrap2(c, args[i]))
	}
	argsHash := types.WrapHash(entries)
	result, err := ga.client.InvokeAction(c, &fsmpb.ActionInvocation{Id: ga.id, Arguments: misc.ToPBData(argsHash).GetHashValue()})
	if err != nil {
		return nil, err
	}
	rh := misc.FromPBData(&datapb.Data{Kind: &datapb.Data_HashValue{result}}).(*types.HashValue)
	rm := make(map[string]reflect.Value, rh.Len())
	rf := c.Reflector()
	rh.EachPair(func(k, v eval.PValue) {
		if vt, ok := rf.ReflectType(v.Type()); ok {
			rm[k.String()] = rf.Reflect(v, vt)
		} else {
			panic(eval.Error(c, fsm.GENESIS_UNABLE_TO_REFLECT_TYPE, issue.H{`type`: vt}))
		}
	})
	return rm, nil
}

func convertParams(c eval.Context, parameters []*fsmpb.Parameter) []*fsm.Parameter {
	fp := make([]*fsm.Parameter, len(parameters))
	for i, p := range parameters {
		fp[i] = fsm.NewParameter(p.Name, c.ParseType2(p.Type))
	}
	return fp
}
