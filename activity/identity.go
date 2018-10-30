package activity

type Identity interface {
	// Exists returns true if a resource corresponding to the unique identifier exists
	// in the managed infrastructure
	Exists(identity string) bool
}
