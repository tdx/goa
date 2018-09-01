package stdlib

//
// Tracer is the interface that defines the Event method
//
type Tracer interface {
	//
	// Event handles tracer events
	//
	Event(events ...Term)
}

//
// TracerFunc type is an adapter to allow to use of ordinary func as tracer
//
type TracerFunc func(events ...Term)

//
// Event implements TracerFunc
//
func (f TracerFunc) Event(events ...Term) {
	f(events...)
}

//
// DefaultTracer skips events
//
func emptyTracer() TracerFunc {
	return func(events ...Term) {}
}

//
// Middleware
//

//
// TracerChain makes a chain of tracers call
//
func TracerChain(outer Tracer, rest ...Tracer) Tracer {

	return func(t Tracer) Tracer {
		//
		// t (emptyTracer()) will be last tracer
		//
		topIndex := len(rest) - 1
		for i := range rest {
			t = middleware(rest[topIndex-i], t)
		}
		return middleware(outer, t)
	}(emptyTracer())
}

func middleware(current, next Tracer) Tracer {
	mwFunc := func(mwNext Tracer) Tracer {
		traceFunc := func(events ...Term) {
			current.Event(events...)
			mwNext.Event(events...)
		}
		return TracerFunc(traceFunc)
	}
	return mwFunc(next)
}
