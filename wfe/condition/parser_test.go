package condition

import (
	"fmt"
	"github.com/puppetlabs/go-evaluator/eval"

	// Ensure pcore initialization
	_ "github.com/puppetlabs/go-evaluator/pcore"
)

func ExampleParse() {
	err := eval.Puppet.Do(func(ctx eval.Context) error {
		c := Parse("hello")
		fmt.Println(c)
		return nil
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	// Output: hello
}

func ExampleParse_and() {
	err := eval.Puppet.Do(func(ctx eval.Context) error {
		c := Parse("hello and goodbye")
		fmt.Println(c)
		return nil
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	// Output: hello and goodbye
}

func ExampleParse_not() {
	err := eval.Puppet.Do(func(ctx eval.Context) error {
		c := Parse("!(hello and goodbye)")
		fmt.Println(c)
		return nil
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	// Output: !(hello and goodbye)
}

func ExampleParse_or() {
	err := eval.Puppet.Do(func(ctx eval.Context) error {
		c := Parse("greeting and (hello or goodbye)")
		fmt.Println(c)
		return nil
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	// Output: greeting and (hello or goodbye)
}
