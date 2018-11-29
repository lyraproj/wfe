package wfe

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-evaluator/types"
	"github.com/puppetlabs/go-hiera/lookup"
	"github.com/puppetlabs/go-issues/issue"
	"github.com/puppetlabs/go-puppet-dsl-workflow/puppet"
	"github.com/puppetlabs/go-servicesdk/grpc"
	"github.com/puppetlabs/go-servicesdk/service"
	"github.com/puppetlabs/go-servicesdk/serviceapi"
	"github.com/puppetlabs/go-servicesdk/wfapi"
	"os"
	"os/exec"

	// Ensure Pcore and lookup are initialized
	_ "github.com/puppetlabs/go-evaluator/pcore"
	_ "github.com/puppetlabs/go-hiera/functions"
	_ "github.com/puppetlabs/go-servicesdk/wf"
	_ "github.com/puppetlabs/go-puppet-dsl-workflow/puppet/functions"
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

func provider(c lookup.ProviderContext, key string, _ eval.OrderedMap) (eval.Value, bool) {
	v, ok := sampleData.Get4(key)
	return v, ok
}

func withSampleService(sf func(eval.Context)) {
	eval.Puppet.Set(`tasks`, types.Boolean_TRUE)
	eval.Puppet.Set(`workflow`, types.Boolean_TRUE)
	lookup.DoWithParent(context.Background(), provider, func(ctx eval.Context) {
		// Command to start plug-in and read a given manifest
		testRoot := `../../go-puppet-dsl-workflow`
		cmd := exec.Command("go", "run", testRoot + "/main/main.go", testRoot + `/puppet/testdata/attach.pp`)

		// Logger that prints JSON on Stderr
		logger := hclog.New(&hclog.LoggerOptions{
			Level:      hclog.Debug,
			Output:     os.Stderr,
			JSONFormat: true,
		})

		server, err := grpc.Load(cmd, logger)
		defer func() {
			// Ensure that plug-ins die when we're done.
			plugin.CleanupClients()
		}()

		if err == nil {
			ctx.Scope().Set(ActivityContextKey, types.SingletonHash2(`operation`, types.WrapInteger(int64(wfapi.Upsert))))
			withTestLoader(ctx, server, sf)
		} else {
			fmt.Println(err)
		}
	})
}

func withSampleLocalService(sf func(eval.Context)) {
	eval.Puppet.Set(`tasks`, types.Boolean_TRUE)
	eval.Puppet.Set(`workflow`, types.Boolean_TRUE)
	lookup.DoWithParent(context.Background(), provider, func(ctx eval.Context) {
		testRoot := `../../go-puppet-dsl-workflow`
		service := puppet.CreateService(ctx, `Puppet`, testRoot + `/puppet/testdata/attach.pp`)
		ctx.Scope().Set(ActivityContextKey, types.SingletonHash2(`operation`, types.WrapInteger(int64(wfapi.Upsert))))
		withTestLoader(ctx, service, sf)
	})
}

func ExampleRemoteActivity() {
	eval.Puppet.Set(`tasks`, types.Boolean_TRUE)
	eval.Puppet.Set(`workflow`, types.Boolean_TRUE)
	withSampleService(func(c eval.Context) {
		wf := CreateActivity(GetDefinition(c, eval.NewTypedName(eval.NsDefinition, `attach`)))
		result := wf.Run(c, eval.EMPTY_MAP)
		fmt.Println(eval.ToPrettyString(result))
	})
	// Output:
	// {
	//   'vpc_id' => 'external-vpc-id',
	//   'subnet_id' => 'external-subnet-id',
	//   'internet_gateway_id' => 'external-internet_gateway_id',
	//   'nodes' => {
	//     'external-instance-id' => {
	//       'public_ip' => '192.168.0.20',
	//       'private_ip' => '192.168.1.20'
	//     },
	//     'external-instance-id' => {
	//       'public_ip' => '192.168.0.20',
	//       'private_ip' => '192.168.1.20'
	//     },
	//     'external-instance-id' => {
	//       'public_ip' => '192.168.0.20',
	//       'private_ip' => '192.168.1.20'
	//     },
	//     'external-instance-id' => {
	//       'public_ip' => '192.168.0.20',
	//       'private_ip' => '192.168.1.20'
	//     },
	//     'external-instance-id' => {
	//       'public_ip' => '192.168.0.20',
	//       'private_ip' => '192.168.1.20'
	//     }
	//   }
	// }
}

var mockServiceId = eval.NewTypedName(eval.NsService, `MockService`)

func createMockIdentityService() (s serviceapi.Service) {
	eval.Puppet.Do(func(c eval.Context) {
		sb := service.NewServerBuilder(c, `Mock::Identity`)
		sb.RegisterAPI(serviceapi.IdentityName, &mockIdentity{})
		s = sb.Server()
	})
	return
}

var mockIdentityType eval.Type

type mockIdentity struct {}

func (m *mockIdentity) Associate(internalID, externalID string) error {
	return nil
}

func (m *mockIdentity) GetExternal(internalID string) (externalID string, ok bool, err error) {
	return `ext_` + internalID, true, nil
}

func (m *mockIdentity) GetInternal(externalID string) (internalID string, ok bool, err error) {
	return externalID[4:], true, nil
}

func (m *mockIdentity) RemoveExternal(externalID string) error {
	return nil
}

func (m *mockIdentity) RemoveInternal(internalID string) error {
	return nil
}

func withTestLoader(c eval.Context, s serviceapi.Service, doer func(eval.Context)) {
	tl := eval.NewParentedLoader(c.Loader())
	c.DoWithLoader(tl, func() {
		addService(c, tl, s)
		addService(c, tl, createMockIdentityService())
		doer(c)
	})
}

func addService(c eval.Context, tl eval.DefiningLoader, s serviceapi.Service) {
	tl.SetEntry(s.Identifier(c), eval.NewLoaderEntry(s, nil))
	_, md := s.Metadata(c)
	for _, def := range md {
		tl.SetEntry(def.Identifier(), eval.NewLoaderEntry(def, nil))
		if handlerFor, ok := def.Properties().Get4(`handler_for`); ok {
			hn := eval.NewTypedName(eval.NsHandler, handlerFor.(issue.Named).Name())
			tl.SetEntry(hn, eval.NewLoaderEntry(def, nil))
		}
	}
}

func ExampleActivity() {
	eval.Puppet.Do(func(c eval.Context) {
		wf := wfapi.NewWorkflow(c, func(wb wfapi.WorkflowBuilder) {
			wb.Name(`Wftest`)
			wb.Output(wb.Parameter(`r`, `String`))
			wb.Stateless(func(sb wfapi.StatelessBuilder) {
				sb.Name(`A`)
				sb.Output(sb.Parameter(`a`, `String`), sb.Parameter(`b`, `Integer`))
				sb.Doer(func(in map[string]interface{}) map[string]interface{} {
					return map[string]interface{}{`a`: `hello`, `b`: 4}
				})
			})
			wb.Stateless(func(sb wfapi.StatelessBuilder) {
				sb.Name(`B1`)
				sb.Input(sb.Parameter(`a`, `String`), sb.Parameter(`b`, `Integer`))
				sb.Output(sb.Parameter(`c`, `String`), sb.Parameter(`d`, `Integer`))
				sb.Doer(func(in map[string]interface{}) map[string]interface{} {
					return map[string]interface{}{`c`: in[`a`].(string) + ` world`, `d`: in[`b`].(int64) + 4}
				})
			})
			wb.Stateless(func(sb wfapi.StatelessBuilder) {
				sb.Name(`B2`)
				sb.Input(sb.Parameter(`a`, `String`), sb.Parameter(`b`, `Integer`))
				sb.Output(sb.Parameter(`e`, `String`), sb.Parameter(`f`, `Integer`))
				sb.Doer(func(in map[string]interface{}) map[string]interface{} {
					return map[string]interface{}{`e`: in[`a`].(string) + ` earth`, `f`: in[`b`].(int64) + 8}
				})
			})
			wb.Stateless(func(sb wfapi.StatelessBuilder) {
				sb.Name(`C`)
				sb.Input(sb.Parameter(`c`, `String`), sb.Parameter(`d`, `Integer`), sb.Parameter(`e`, `String`), sb.Parameter(`f`, `Integer`))
				sb.Output(sb.Parameter(`r`, `String`))
				sb.Doer(func(in map[string]interface{}) map[string]string {
					return map[string]string{`r`: fmt.Sprintf("%s, %d, %s, %d\n", in[`c`], in[`d`], in[`e`], in[`f`])}
				})
			})
		})

		sb := service.NewServerBuilder(c, `My::Plugin`)
		sb.RegisterActivity(wf)
		withTestLoader(c, sb.Server(), func(c eval.Context) {
			wf := CreateActivity(GetDefinition(c, eval.NewTypedName(eval.NsDefinition, `wftest`)))
			result := wf.Run(c, eval.EMPTY_MAP)
			fmt.Println(result.Get5(`r`, eval.EMPTY_STRING))
		})
	})

	// Output: hello world, 8, hello earth, 12
}

/*
func ExampleActivity_failValiation() {
	type AB struct {
		A string
		B int64
	}

	type CD struct {
		C string
		D int64
	}

	type ABCDE struct {
		A string
		B int64
		C string
		D int64
		E string
	}

	type F struct {
		F string
	}

	err := eval.Puppet.Try(func(ctx eval.Context) error {

		// Run actions in process by adding actions directly to the actor server
		as := wfe.NewWorkflow(`wftest`, []eval.Parameter{}, []eval.Parameter{}, nil,

			Activity(ctx, "a", func() (*AB, error) {
				return &AB{`hello`, 4}, nil
			}),

			Activity(ctx, "b", func(in *AB) (*CD, error) {
				return &CD{in.A + ` world`, in.B + 4}, nil
			}),

			Activity(ctx, "c", func(in *ABCDE) (*F, error) {
				return &F{in.A + ` earth`}, nil
			}))

		result := as.Run(ctx, eval.EMPTY_MAP)
		fmt.Println(result.Get5(`r`, eval.EMPTY_STRING))
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}

	// Output: activity 'c' value 'e' is never produced
}

var sampleData = map[string]eval.Value{
	`L`: types.WrapString(`value of L`),
}

func provider(c lookup.Invocation, key string, _ eval.OrderedMap) eval.Value {
	if v, ok := sampleData[key]; ok {
		return v
	}
	c.NotFound()
	return nil
}

func ExampleActivity_lookup() {
	type L struct {
		L string `puppet:"value=>Deferred(lookup,['L'])"`
	}

	type AB struct {
		A string
		B int64
	}

	type CD struct {
		C string
		D int64
	}

	type ABCD struct {
		A string
		B int64
		C string
		D int64
	}

	type F struct {
		F string
	}

	err := lookup.TryWithParent(context.Background(), provider, func(ctx lookup.Context) error {
		// Run actions in process by adding actions directly to the actor server
		wf := wfe.NewWorkflow(`wftest`, []eval.Parameter{}, MakeParams(ctx, `wftest`, &F{}), nil,

			Activity(ctx, "a", func(in *L) (*AB, error) {
				return &AB{in.L, 4}, nil
			}),

			Activity(ctx, "b", func(in *AB) (*CD, error) {
				return &CD{strings.ToLower(in.A), in.B + 4}, nil
			}),

			Activity(ctx, "c", func(in *ABCD) (*F, error) {
				return &F{fmt.Sprintf(`%s %d, %s %d`, in.A, in.B, in.C, in.D)}, nil
			}))

		result := wf.Run(ctx, eval.EMPTY_MAP)
		fmt.Println(result.Get5(`f`, eval.EMPTY_STRING))
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}

	// Output: value of L 4, value of l 8
}
*/
