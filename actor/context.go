package actor

import (
	"fmt"
	"github.com/puppetlabs/data-protobuf/datapb"
	"github.com/puppetlabs/go-fsm/fsm"
	"github.com/puppetlabs/go-fsm/fsmpb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"reflect"
)

type Context interface {
	fsm.Context

	RegisterServer(network, address string) error
}

type actor struct {
	context.Context
	actions []fsm.Action
}

func (s *actor) GetActions(ctx context.Context, request *fsmpb.ActionsRequest) (*fsmpb.ActionsResponse, error) {
	aa := make([]*fsmpb.Action, len(s.actions))
	for i, a := range s.actions {
		aa[i] = &fsmpb.Action{Id: int64(i), Name: a.Name(), Consumes: a.Consumes(), Produces: a.Produces()}
	}
	return &fsmpb.ActionsResponse{Actions: aa}, nil
}

func (s *actor) InvokeAction(ctx context.Context, in *fsmpb.ActionInvocation) (*datapb.DataHash, error) {
	id := int(in.Id)
	if id < 0 || id >= len(s.actions) {
		return nil, fmt.Errorf("no action with ID %d", id)
	}

	ep := in.Arguments.Entries
	rm := make(map[string]reflect.Value, len(ep))
	for _, p := range ep {
		arg, err := datapb.FromData(p.Value)
		if err != nil {
			return nil, err
		}
		rm[p.Key] = arg
	}
	cr := s.actions[id].Call(s, rm)
	ep = make([]*datapb.DataEntry, 0, len(cr))
	for k, v := range cr {
		rv, err := datapb.ToData(v)
		if err != nil {
			return nil, err
		}
		ep = append(ep, &datapb.DataEntry{Key: k, Value: rv})
	}
	return &datapb.DataHash{Entries: ep}, nil
}

func (s *actor) Action(name string, function interface{}) {
	s.actions = append(s.actions, fsm.NewGoAction(name, function))
}

func NewContext() Context {
	return &actor{Context: context.Background(), actions: make([]fsm.Action, 0)}
}

func (as *actor) RegisterServer(network, address string) error {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		return err
	}
	s := grpc.NewServer()
	fsmpb.RegisterActorServer(s, as)

	// Register reflection service on gRPC actor.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		return err
	}
	return nil
}
