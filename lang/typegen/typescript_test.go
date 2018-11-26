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

func ExampleGenerator_GenerateTypes() {
	type Address struct {
		Street string
		Zip    string `puppet:"name=>zip_code"`
	}
	type Person struct {
		Name    string
		Gender  string `puppet:"type=>Enum[male,female,other]"`
		Address *Address
	}
	type ExtendedPerson struct {
		Person
		Age    int  `puppet:"type=>Optional[Integer],value=>undef"`
		Active bool `puppet:"name=>enabled"`
	}

	c := eval.Puppet.RootContext()

	// Create a TypeSet from a list of Go structs
	typeSet := c.Reflector().TypeSetFromReflect(`My::Own`, semver.MustParseVersion(`1.0.0`), nil,
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
	//
	//       __pvalue() : {[s: string]: any} {
	//         let ih: {[s: string]: any} = {};
	//         ih['street'] = this.street;
	//         ih['zip_code'] = this.zip_code;
	//         return ih;
	//       }
  //
	//       __ptype() : string {
	//         return 'My::Own::Address';
	//       }
	//     }
	//
	//     export class Person {
	//       readonly name: string;
	//       readonly gender: 'male' | 'female' | 'other';
	//       readonly address: Address;
	//
	//       constructor({
	//           name,
	//           gender,
	//           address
	//         }: {
	//           name: string,
	//           gender: 'male' | 'female' | 'other',
	//           address: Address
	//         }) {
	//         this.name = name;
	//         this.gender = gender;
	//         this.address = address;
	//       }
	//
	//       __pvalue() : {[s: string]: any} {
	//         let ih: {[s: string]: any} = {};
	//         ih['name'] = this.name;
	//         ih['gender'] = this.gender;
	//         ih['address'] = this.address;
	//         return ih;
	//       }
  //
	//       __ptype() : string {
	//         return 'My::Own::Person';
	//       }
	//     }
	//
	//     export class ExtendedPerson extends Person {
	//       readonly enabled: boolean;
	//       readonly age: number | null;
	//
	//       constructor({
	//           name,
	//           gender,
	//           address,
	//           enabled,
	//           age = null
	//         }: {
	//           name: string,
	//           gender: 'male' | 'female' | 'other',
	//           address: Address,
	//           enabled: boolean,
	//           age?: number | null
	//         }) {
	//         super({name: name, gender: gender, address: address});
	//         this.enabled = enabled;
	//         this.age = age;
	//       }
  //
	//       __pvalue() : {[s: string]: any} {
	//         let ih = super.__pvalue();
	//         ih['enabled'] = this.enabled;
	//         if(this.age !== null)
	//           ih['age'] = this.age;
	//         return ih;
	//       }
  //
	//       __ptype() : string {
	//         return 'My::Own::ExtendedPerson';
	//       }
	//     }
	//   }
	// }
}