package puppet

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-fsm/wfe"

	// Ensure pcore is initialized
	_ "github.com/puppetlabs/go-evaluator/pcore"
)

func init() {
	eval.NewGoFunction(`apply`,
		func(d eval.Dispatch) {
			d.Param(`String`)
			d.Param(`Object`)
			d.Function(func(c eval.Context, args []eval.Value) eval.Value {
				return wfe.Apply(c, args[0].String(), args[1].(eval.PuppetObject))
			})
		})
}
