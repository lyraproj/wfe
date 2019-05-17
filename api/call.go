package api

type Call interface {
	Step

	CalledStep() Step
}
