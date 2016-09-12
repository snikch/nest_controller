package controller

// Event represents a change in state that has occurred.
type Event struct {
	Entity string
	Value  string
}

// EventHandler defines an interface for handling events.
type EventHandler interface {
	Push(Event) error
	Name() string
}
