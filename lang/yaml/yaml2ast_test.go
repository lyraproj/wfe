package yaml2ast

import (
	"context"
	"fmt"
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/types"
	"github.com/puppetlabs/go-fsm/api"
	"github.com/puppetlabs/go-hiera/lookup"
	"github.com/puppetlabs/go-issues/issue"
	"io/ioutil"
	"os"

	// Ensure that pcore is initialized
	_ "github.com/puppetlabs/go-evaluator/pcore"
)

var sampleData = eval.Wrap(nil, map[string]interface{}{
	`aws`: map[string]interface{}{
		`region`:  `eu-west-1`,
		`keyname`: `aws-key-name`,
		`tags`: map[string]string{
			`created_by`: `john.mccabe@puppet.com`,
			`department`: `engineering`,
			`project`:    `incubator`,
			`lifetime`:   `1h`,
		},
		`instance`: map[string]interface{}{
			`count`: 5,
		}}}).(*types.HashValue)

func provider(c lookup.Context, key string, _ eval.OrderedMap) eval.Value {
	if v, ok := sampleData.Get4(key); ok {
		return v
	}
	c.NotFound()
	return nil
}

func ExampleActivity() {
	eval.Puppet.Set(`tasks`, types.Boolean_TRUE)
	eval.Puppet.Set(`workflow`, types.Boolean_TRUE)
	err := lookup.DoWithParent(context.Background(), provider, func(ctx lookup.Context) error {
		workflowName := `attach`
		path := `testdata/` + workflowName + `.yaml`
		content, err := ioutil.ReadFile(path)
		if err != nil {
			panic(eval.Error(eval.EVAL_UNABLE_TO_READ_FILE, issue.H{`path`: path, `detail`: err.Error()}))
		}
		ast := YamlToAST(ctx, path, content)
		ctx.AddDefinitions(ast)
		_, err = ctx.Evaluator().Evaluate(ctx, ast)
		if err != nil {
			return err
		}

		wf, ok := eval.Load(ctx, eval.NewTypedName(eval.WORKFLOW, workflowName))
		if !ok {
			return fmt.Errorf(`%s did not define workflow %s`, path, workflowName)
		}
		rf := wf.(api.Activity).Run(ctx, eval.EMPTY_MAP)
		rf.ToString(os.Stdout, eval.PRETTY, nil)
		return nil
	})

	if err != nil {
		fmt.Println(err.Error())
	}

	// Output:
	// {
	//   'vpc_id' => 'FAKED_VPC_ID',
	//   'subnet_id' => 'FAKED_SUBNET_ID',
	//   'internet_gateway_id' => 'FAKED_GATEWAY_ID',
	//   'nodes' => {
	//     '0' => {
	//       'public_ip' => 'FAKED_PUBLIC_IP',
	//       'private_ip' => 'FAKED_PRIVATE_IP'
	//     },
	//     '1' => {
	//       'public_ip' => 'FAKED_PUBLIC_IP',
	//       'private_ip' => 'FAKED_PRIVATE_IP'
	//     },
	//     '2' => {
	//       'public_ip' => 'FAKED_PUBLIC_IP',
	//       'private_ip' => 'FAKED_PRIVATE_IP'
	//     },
	//     '3' => {
	//       'public_ip' => 'FAKED_PUBLIC_IP',
	//       'private_ip' => 'FAKED_PRIVATE_IP'
	//     },
	//     '4' => {
	//       'public_ip' => 'FAKED_PUBLIC_IP',
	//       'private_ip' => 'FAKED_PRIVATE_IP'
	//     }
	//   }
	// }
	//
}
