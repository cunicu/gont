package gont

type Node interface {
	Teardown() error

	// Getters
	Name() string
	Base() *BaseNode
}
