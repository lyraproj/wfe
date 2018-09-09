package api

type Workflow interface {
	Activity

	Activities() []Activity
}
