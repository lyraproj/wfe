package server

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-plugin"
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/proto"
	"github.com/puppetlabs/go-fsm/api"
	"github.com/puppetlabs/go-fsm/lang/rpc/fsmpb"
	"github.com/puppetlabs/go-fsm/lang/rpc/shared"
	"google.golang.org/grpc"
	"io"
	"net/rpc"
	"sort"

	// Ensure that Puppet is initialized
	_ "github.com/puppetlabs/go-evaluator/pcore"
)

type ActorsPlugin struct {
	actors map[string]api.Workflow
}

func NewActorsPlugin(actors ...api.Workflow) *ActorsPlugin {
	am := make(map[string]api.Workflow, len(actors))
	for _, a := range actors {
		am[a.Name()] = a
	}
	return &ActorsPlugin{am}
}

func (a *ActorsPlugin) GetActor(name string) api.Workflow {
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
	actor := s.impl.GetActor(ar.GetName())
	return &fsmpb.Actor{
		Actions: convertToPbActions(actor.GetActivities()),
		Input:   shared.ConvertToPbParams(actor.Input()),
		Output:  shared.ConvertToPbParams(actor.Output()),
	}, nil
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
			return fmt.Errorf(`InvokeActivity expects a 3 element array`)
		}
		actorName := da.Values[0].GetStringValue()
		actionName := da.Values[1].GetStringValue()

		rm := proto.FromPBData(da.Values[2]).(eval.OrderedMap)
		rv := s.impl.GetActor(actorName).InvokeActivity(actionName, rm, NewGenesis(eval.Puppet.RootContext(), stream))
		result := proto.ToPBData(rv)

		err = stream.Send(&fsmpb.Message{Id: shared.InvokeActionId, Value: result})
		if err != nil {
			return err
		}
	}
}

func convertToPbActions(actions map[string]api.Activity) []*fsmpb.Action {
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
