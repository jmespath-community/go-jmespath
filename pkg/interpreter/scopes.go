package interpreter

type Scope interface {
	GetValue(string) (interface{}, bool)
	With(data map[string]interface{}) Scope
}

type scope struct {
	inner Scope
	data  map[string]interface{}
}

// newScope creates a new instance of JMESPath scope.
func newScope(data map[string]interface{}) Scope {
	return scope{
		data: data,
	}
}

func (s scope) GetValue(identifier string) (interface{}, bool) {
	if s.data != nil {
		if item, ok := s.data[identifier]; ok {
			return item, true
		}
	}
	if s.inner != nil {
		return s.inner.GetValue(identifier)
	}
	return nil, false
}

func (s scope) With(data map[string]interface{}) Scope {
	return scope{
		data:  data,
		inner: s,
	}
}
