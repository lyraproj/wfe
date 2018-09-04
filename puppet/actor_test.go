package puppet

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/types"
	"io/ioutil"
	"github.com/puppetlabs/go-issues/issue"
	"fmt"

	// Ensure Pcore is initialized
	_ "github.com/puppetlabs/go-evaluator/pcore"
	"github.com/puppetlabs/go-fsm/api"
	"github.com/puppetlabs/go-fsm/fsm"
	"reflect"
)

func ExampleActor() {
	eval.Puppet.Set(`tasks`, types.Boolean_TRUE)
	eval.Puppet.Set(`actors`, types.Boolean_TRUE)
	err := eval.Puppet.Do(func(ctx eval.Context) error {
		actorName := `attach`
		path := `testdata/` + actorName + `.pp`
		content, err := ioutil.ReadFile(path)
		if err != nil {
			panic(eval.Error(ctx, eval.EVAL_UNABLE_TO_READ_FILE, issue.H{`path`: path, `detail`: err.Error()}))
		}
		ast := ctx.ParseAndValidate(path, string(content), false)
		ctx.AddDefinitions(ast)
		_, err = ctx.Evaluator().Evaluate(ctx, ast)
		if err != nil {
			return err
		}

		actor, ok := eval.Load(ctx, eval.NewTypedName(eval.ACTOR, actorName))
		if !ok {
			return fmt.Errorf(`%s did not define actor %s`, path, actorName)
		}
		actorServer := fsm.NewActorServer2(ctx, actor.(api.Actor))
		err = actorServer.Validate()
		if err != nil {
			return err
		}
		rf := actorServer.Call(nil, map[string]reflect.Value{})
		fmt.Println(rf)
		return nil
	})

	if err != nil {
		fmt.Println(err.Error())
	}

	// Output: map[vpc_id:FAKED_VPC_ID subnet_id:FAKED_SUBNET_ID internet_gateway_id:FAKED_GATEWAY_ID]
}
