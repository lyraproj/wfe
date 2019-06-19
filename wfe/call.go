package wfe

type Call interface {
	Step

	// CalledStep returns the Step that this step will call
	CalledStep() Step
}
