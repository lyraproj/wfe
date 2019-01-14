package yaml2ast

import (
	"context"
	"fmt"
	"github.com/lyraproj/hiera/lookup"
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/puppet-evaluator/eval"
	"github.com/lyraproj/puppet-evaluator/types"
	"github.com/lyraproj/wfe/api"
	"io/ioutil"
	"os"

	// Ensure that pcore is initialized
	_ "github.com/lyraproj/puppet-evaluator/pcore"
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

func provider(c lookup.ProviderContext, key string, _ map[string]eval.Value) (eval.Value, bool) {
	v, ok := sampleData.Get4(key)
	return v, ok
}

func ExampleActivity() {
	eval.Puppet.Set(`tasks`, types.Boolean_TRUE)
	eval.Puppet.Set(`workflow`, types.Boolean_TRUE)
	err := lookup.TryWithParent(context.Background(), provider, nil, func(ctx eval.Context) error {
		workflowName := `attach`
		path := `testdata/` + workflowName + `.yaml`
		content, err := ioutil.ReadFile(path)
		if err != nil {
			panic(eval.Error(eval.EVAL_UNABLE_TO_READ_FILE, issue.H{`path`: path, `detail`: err.Error()}))
		}
		ast := YamlToAST(ctx, path, content)
		ctx.AddDefinitions(ast)
		_, err = eval.TopEvaluate(ctx, ast)
		if err != nil {
			return err
		}

		wf, ok := eval.Load(ctx, eval.NewTypedName(eval.NsActivity, workflowName))
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
}
