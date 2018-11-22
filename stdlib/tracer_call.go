package stdlib

import (
	"time"
)

//
// Call is a predefined tracer event for GenProc. Fired before call.
//
type Call struct {
	Pid  *Pid
	Time *time.Time
	Tag  string
	Arg  Term
}

//
// CallResult is a predefined tracer event for GenProc. Fired after call.
//
type CallResult struct {
	Call
	Result   Term
	Duration time.Duration
}

//
// TraceCall sends event to tracer as message of type *Call
//
func TraceCall(t Tracer, pid *Pid, tag string, arg Term) *time.Time {

	if t == nil {
		return nil
	}

	now := time.Now()
	t.Event(&Call{pid, &now, tag, arg})
	return &now
}

//
// TraceCallResult sends event to tracer as message of type *CallResult
//
func TraceCallResult(
	t Tracer, pid *Pid, start *time.Time, tag string, arg, result Term) {

	if t == nil || start == nil {
		return
	}

	now := time.Now()
	t.Event(&CallResult{
		Call{pid, &now, tag, arg},
		result,
		now.Sub(*start),
	})
}
