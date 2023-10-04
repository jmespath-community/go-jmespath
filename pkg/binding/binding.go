package binding

// Binding in the interface representing a let expression binding.
// You can get the value of the binding by calling the `Value` method.
type Binding interface {
	// Get returns the value bound for a given name.
	Value() (interface{}, error)
}

type binding struct {
	value interface{}
}

func (b *binding) Value() (interface{}, error) {
	return b.value, nil
}

func NewBinding(value interface{}) Binding {
	return &binding{
		value: value,
	}
}
