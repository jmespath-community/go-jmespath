package interpreter

type Option func(options) options

type options struct {
	functionCaller FunctionCaller
}

func WithFunctionCaller(functionCaller FunctionCaller) Option {
	return func(o options) options {
		o.functionCaller = functionCaller
		return o
	}
}
