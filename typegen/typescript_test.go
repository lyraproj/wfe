package typegen

import (
	"github.com/puppetlabs/go-evaluator/eval"
	"github.com/puppetlabs/go-semver/semver"
	"reflect"
	"bytes"
	// Initialize pcore
	_ "github.com/puppetlabs/go-evaluator/pcore"
	"fmt"
)

func ExampleReflector_typeSetFromReflect() {
	type Address struct {
		Street string
		Zip    string `puppet:"name=>zip_code"`
	}
	type Person struct {
		Name    string
		Address *Address
	}
	type ExtendedPerson struct {
		Person
		Age    int  `puppet:"type=>Optional[Integer],value=>undef"`
		Active bool `puppet:"name=>enabled"`
	}

	c := eval.Puppet.RootContext()

	// Create a TypeSet from a list of Go structs
	typeSet := c.Reflector().TypeSetFromReflect(`My`, semver.MustParseVersion(`1.0.0`),
		reflect.TypeOf(&Address{}), reflect.TypeOf(&Person{}), reflect.TypeOf(&ExtendedPerson{}))

	// Make the types known to the current loader
	c.AddTypes(typeSet)

	bld := bytes.NewBufferString(``)
	GenerateTypes(c, typeSet, []string{}, 0, bld)
	fmt.Println(bld.String())

	// Output: Fooo
}