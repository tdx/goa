package stdlib

import (
	"fmt"
	"time"
)

//
// Example tests
//
func Example_traceFunc() {
	tr := traceToTracerFunc()
	tr.Event("some event")
	// Output: traceToTracerFunc called
}

func Example_traceFromVar() {
	var tr TracerFunc = func(events ...Term) {
		fmt.Println(events[0])
	}
	tr.Event("some event")
	// Output: some event
}

func Example_traceInterfaceImplementation() {
	obj := new(someTracer)
	tr := TracerFunc(obj.Trace)
	tr.Event("some event")
	// Output: tracer with receiver
}

func ExampleTracerChain() {

	//
	// tracer func
	//
	var tracerFromVar TracerFunc = func(events ...Term) {
		fmt.Println("tracerFromVar called")
	}

	//
	// tracer object
	//
	someTr := new(someTracer)

	tr := TracerChain(
		traceToTracerFunc(),
		tracerFromVar,
		TracerFunc(someTr.Trace),
	)
	now := time.Now()
	call := Call{nil, &now, "test2", 124}
	tr.Event(&call)
	tr.Event(&CallResult{call, "aaa", time.Duration(5) * time.Microsecond})
	// Output: traceToTracerFunc called
	// tracerFromVar called
	// tracer with receiver
	// traceToTracerFunc called
	// tracerFromVar called
	// tracer with receiver
}

//
// Locals
//
type someTracer struct {
}

func (t *someTracer) Trace(events ...Term) {
	fmt.Println("tracer with receiver")
}

func traceToTracerFunc() TracerFunc {

	return func(events ...Term) {
		fmt.Println("traceToTracerFunc called")
	}
}
