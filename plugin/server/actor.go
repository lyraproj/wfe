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
	"github.com/puppetlabs/go-fsm/plugin/shared"
	"sort"
)

type ActorsPlugin struct {
	actors map[string]api.Actor
}

func NewActorsPlugin(actors map[string]api.Actor) *ActorsPlugin {
	return &ActorsPlugin{actors}
}

func (a *ActorsPlugin) GetActor(name string) api.Actor {
	actor, found := a.actors[name]
	if !found {
		panic(fmt.Errorf("no such actor '%s'", name))
	}
	return actor
}

func (a *ActorsPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return nil, fmt.Errorf(`%T has no server implementation for rpc`, a)
}

func (a *ActorsPlugin) Client(*plugin.MuxBroker, *rpc.Client) (interface{}, error) {
	return nil, fmt.Errorf(`%T has no client implementation for rpc`, a)
}

func (a *ActorsPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	fsmpb.RegisterActorsServer(s, &GRPCServer{impl: a})
	return nil
}

func (a *ActorsPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return nil, fmt.Errorf(`%T has no client implementation for grpc`, a)
}

type GRPCServer struct {
	impl *ActorsPlugin
}

func (s *GRPCServer) GetActor(ctx context.Context, ar *fsmpb.ActorRequest) (*fsmpb.Actor, error) {
	return &fsmpb.Actor{Actions: convertToPbActions(s.impl.GetActor(ar.GetName()).GetActions())}, nil
}

func (s *GRPCServer) InvokeAction(stream fsmpb.Actors_InvokeActionServer) error {
	for {
		in, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if in.GetId() != shared.InvokeActionId {
			continue
		}
		da := in.GetValue().GetArrayValue()
		if da == nil || len(da.Values) != 3 {
			return fmt.Errorf(`InvokeAction expects a 3 element array`)
		}
		actorName := da.Values[0].GetStringValue()
		actionName := da.Values[1].GetStringValue()

		ep := da.Values[2].GetHashValue().Entries
		rm := make(map[string]reflect.Value, len(ep))
		for _, p := range ep {
			arg, err := datapb.FromData(p.Value)
			if err != nil {
				panic(err)
			}
			rm[p.Key] = arg
		}

		rv := s.impl.GetActor(actorName).InvokeAction(actionName, rm, NewGenesis(stream))

		ep = make([]*datapb.DataEntry, 0, len(rv))
		for k, v := range rv {
			rv, err := datapb.ToData(v)
			if err != nil {
				panic(err)
			}
			ep = append(ep, &datapb.DataEntry{Key: k, Value: rv})
		}
		result := &datapb.Data{Kind: &datapb.Data_HashValue{HashValue: &datapb.DataHash{Entries: ep}}}

		err = stream.Send(&fsmpb.Message{Id: shared.InvokeActionId, Value: result})
		if err != nil {
			return err
		}
	}
}

func convertToPbActions(actions map[string]api.Action) []*fsmpb.Action {
	ps := make([]*fsmpb.Action, 0, len(actions))
	for _, p := range actions {
		ps = append(ps, &fsmpb.Action{Name: p.Name(), Input: shared.ConvertToPbParams(p.Input()), Output: shared.ConvertToPbParams(p.Output())})
	}
	// Send in predictable order (sorted alphabetically on name)
	sort.Slice(ps, func(i, j int) bool {
		return ps[i].Name < ps[j].Name
	})
	return ps
}

