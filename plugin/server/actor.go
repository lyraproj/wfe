package server

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-plugin"
	"github.com/puppetlabs/go-fsm/api"
	"github.com/puppetlabs/go-fsm/fsmpb"
	"google.golang.org/grpc"
	"net/rpc"
	"github.com/puppetlabs/data-protobuf/datapb"
	"reflect"
	"io"
)

type Actor struct {
	context.Context
	actions []api.Action
}

func NewActor(ctx context.Context) *Actor {
	return &Actor{ctx, make([]api.Action, 0, 7)}
}

func (s *Actor) Action(name string, function interface{}) {
	s.actions = append(s.actions, api.NewGoAction(name, function))
}

func (a *Actor) GetActions() []*fsmpb.Action {
	return convertToPbActions(a.actions)
}

func (a *Actor) InvokeAction(id int64, parameters *datapb.DataHash, genesis api.Genesis) *datapb.DataHash {
	if id < 0 || int(id) >= len(a.actions) {
		panic(fmt.Errorf("no action with ID %d", id))
	}

	ep := parameters.Entries
	rm := make(map[string]reflect.Value, len(ep))
	for _, p := range ep {
		arg, err := datapb.FromData(p.Value)
		if err != nil {
			panic(err)
		}
		rm[p.Key] = arg
	}
	cr := a.actions[id].Call(genesis, rm)
	ep = make([]*datapb.DataEntry, 0, len(cr))
	for k, v := range cr {
		rv, err := datapb.ToData(v)
		if err != nil {
			panic(err)
		}
		ep = append(ep, &datapb.DataEntry{Key: k, Value: rv})
	}
	return &datapb.DataHash{Entries: ep}
}

func convertToPbActions(actions []api.Action) []*fsmpb.Action {
	ps := make([]*fsmpb.Action, len(actions))
	for i, p := range actions {
		ps[i] = &fsmpb.Action{Id: int64(i), Name: p.Name(), Consumes: convertToPbParams(p.Consumes()), Produces: convertToPbParams(p.Produces())}
	}
	return ps
}

func convertToPbParams(params []api.Parameter) []*fsmpb.Parameter {
	ps := make([]*fsmpb.Parameter, len(params))
	for i, p := range params {
		ps[i] = &fsmpb.Parameter{p.Name(), p.Type()}
	}
	return ps
}

func (a *Actor) Server(*plugin.MuxBroker) (interface{}, error) {
	return nil, fmt.Errorf(`%T has no server implementation for rpc`, a)
}

func (a *Actor) Client(*plugin.MuxBroker, *rpc.Client) (interface{}, error) {
	return nil, fmt.Errorf(`%T has no client implementation for rpc`, a)
}

func (a *Actor) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	fsmpb.RegisterActorServer(s, &GRPCServer{impl: a})
	return nil
}

func (a *Actor) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return nil, fmt.Errorf(`%T has no client implementation for grpc`, a)
}

type GRPCServer struct {
	impl *Actor
}

func (s *GRPCServer) GetActions(ctx context.Context, ar *fsmpb.ActionsRequest) (*fsmpb.ActionsResponse, error) {
	return &fsmpb.ActionsResponse{Actions: s.impl.GetActions()}, nil
}

func (s *GRPCServer) InvokeAction(stream fsmpb.Actor_InvokeActionServer) error {
	for {
		in, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		err = stream.Send(&fsmpb.ActionMessage{in.Id, s.impl.InvokeAction(in.Id, in.Arguments, NewGenesis(stream))})
		if err != nil {
			return err
		}
	}
}
