package stdlib

import (
	"time"
)

//
// Timer to send event to pid
//
type Timer struct {
	timer *time.Timer
}

//
// RunTimerFunc is the type for function passed to RunAfter
//
type RunTimerFunc func()

//
// SendAfter returns stoppable timer, after timeoutMs sends data event
// to pid
//
func (pid *Pid) SendAfter(data Term, timeoutMs uint32) *Timer {

	timer := time.AfterFunc(
		time.Duration(timeoutMs)*time.Millisecond,
		func() {
			_ = pid.Send(data)
		},
	)

	return &Timer{timer: timer}
}

//
// RunAfter calls f after timeoutMs
//
func (pid *Pid) RunAfter(f RunTimerFunc, timeoutMs uint32) *Timer {
	timer := time.AfterFunc(
		time.Duration(timeoutMs)*time.Millisecond,
		f,
	)

	return &Timer{timer: timer}
}

//
// Stop stops the timer
//
func (t *Timer) Stop() {
	if t == nil {
		return
	}

	if !t.timer.Stop() {
		select {
		case <-t.timer.C:
		default:
		}
	}
}
