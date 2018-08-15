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
		Sex     string `puppet:"type=>Enum[male,female,other]"`
		Address *Address
	}
	type ExtendedPerson struct {
		Person
		Age    int  `puppet:"type=>Optional[Integer],value=>undef"`
		Active bool `puppet:"name=>enabled"`
	}

	c := eval.Puppet.RootContext()

	// Create a TypeSet from a list of Go structs
	typeSet := c.Reflector().TypeSetFromReflect(`My::Own`, semver.MustParseVersion(`1.0.0`),
		reflect.TypeOf(&Address{}), reflect.TypeOf(&Person{}), reflect.TypeOf(&ExtendedPerson{}))

	// Make the types known to the current loader
	c.AddTypes(typeSet)

	bld := bytes.NewBufferString(``)
	g := NewTsGenerator(c, []string{`Resource`})
	g.GenerateTypes(typeSet, []string{}, 0, bld)
	fmt.Println(bld.String())

	// Output:
	// export namespace My {
	//   export namespace Own {
	//
	//     export class Address {
	//       readonly street: string;
	//       readonly zip_code: string;
	//
	//       constructor({
	//           street,
	//           zip_code
	//         }: {
	//           street: string,
	//           zip_code: string
	//         }) {
	//         this.street = street;
	//         this.zip_code = zip_code;
	//       }
	//     }
	//
	//     export class Person {
	//       readonly name: string;
	//       readonly sex: 'male' | 'female' | 'other';
	//       readonly address: Address;
	//
	//       constructor({
	//           name,
	//           sex,
	//           address
	//         }: {
	//           name: string,
	//           sex: 'male' | 'female' | 'other',
	//           address: Address
	//         }) {
	//         this.name = name;
	//         this.sex = sex;
	//         this.address = address;
	//       }
	//     }
	//
	//     export class ExtendedPerson extends Person {
	//       readonly enabled: boolean;
	//       readonly age: number | null;
	//
	//       constructor({
	//           name,
	//           sex,
	//           address,
	//           enabled,
	//           age = null
	//         }: {
	//           name: string,
	//           sex: 'male' | 'female' | 'other',
	//           address: Address,
	//           enabled: boolean,
	//           age?: number | null
	//         }) {
	//         super({name: name, sex: sex, address: address});
	//         this.enabled = enabled;
	//         this.age = age;
	//       }
	//     }
	//   }
	// }
}