package api

type GoActorBuilder interface {
	// Add a Go function as an action with the given name. The function
	// must take a required Genesis parameter and an optional pointer
	// to a struct as parameters. It must return a pointer to a struct
	//
	// The "consumes" declaration is reflected from the the fields in the struct
	// parameter and the "produces" declaration is reflected from the fields in
	// the returned struct.
	Action(name string, function interface{})
}
