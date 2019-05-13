module github.com/lyraproj/wfe

require (
	github.com/hashicorp/go-hclog v0.8.0
	github.com/lyraproj/issue v0.0.0-20190329160035-8bc10230f995
	github.com/lyraproj/pcore v0.0.0-20190502085713-c95bdae56d68
	github.com/lyraproj/servicesdk v0.0.0-20190508121759-aa1c3c39fdcb
	gonum.org/v1/gonum v0.0.0-20190331200053-3d26580ed485
)

// Remove once lyraproj/issue#7 is merged
replace github.com/lyraproj/issue => github.com/thallgren/issue v0.0.0-20190512160618-668e97752cb0

// Remove once lyraproj/servicesdk#34 is merged
replace github.com/lyraproj/servicesdk => github.com/thallgren/servicesdk v0.0.0-20190513075437-38570cba00d4
