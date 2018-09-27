package stdlib

//
// Timer GenServer - holds timers in gts and send messages at specified time
//

import (
	"errors"
	"time"
)

var (
	timerPid *Pid
)

//
// TimerServerStart starts timer GenServer
//
func TimerServerStart() (err error) {

	opts := NewSpawnOpts().
		WithName("timer_gs").
		WithSpawnOrLocate()
	timerPid, err = GenServerStart(new(tgs), opts)

	return err
}

//
// TimerSendAfter adds one-time timer
//
func TimerSendAfter(timeMs uint32, pid *Pid, msg Term) (Term, error) {

	if timeMs == 0 {
		return nil, errors.New("bad timer time 0 ms")
	}

	timeout := time.Duration(timeMs) * time.Millisecond
	r := &timerAfterReq{timeout, &timerArgs{pid, msg}, time.Now()}
	tref, err := timerPid.Call(r)
	if err != nil {
		return nil, err
	}

	return tref, nil
}

//
// TimerSendInterval adds interval timer
//
func TimerSendInterval(timeMs uint32, pid *Pid, msg Term) (Term, error) {

	if timeMs == 0 {
		return nil, errors.New("bad timer time 0 ms")
	}

	timeout := time.Duration(timeMs) * time.Millisecond
	r := &timerIntervalReq{timeout, &timerArgs{pid, msg}, time.Now(), timeMs}
	tref, err := timerPid.Call(r)
	if err != nil {
		return nil, err
	}

	return tref, nil
}

//
// TimerCancel deletes timer
//
func TimerCancel(tref Term) error {

	_, err := timerPid.Call(tref)
	return err
}

//
// GenServer callbacks
//
func (gs *tgs) Init(args ...Term) Term {

	gs.SetTrapExit(true)
	// gs.SetTracer(TraceToConsole())

	gs.timerTab = NewOrderedSetWith(cmpTimerRef)
	gs.intervalTab = NewSet()

	return GsInitOk
}

func (gs *tgs) HandleCall(req Term, from From) Term {

	switch req := req.(type) {

	case *timerAfterReq:

		sysTime := time.Now()

		when := req.started.Add(req.timeout)
		tref := &timerRef{when, MakeRef(), 0, req.started}
		gs.timerTab.Insert(tref, req.op)

		return gs.CallReplyTimeout(tref, gs.timeout(sysTime))

	case *timerIntervalReq:

		gs.Link(req.op.pid)

		sysTime := time.Now()
		iref := MakeRef()

		when := req.started.Add(req.timeout)
		tref := &timerRef{when, MakeRef(), req.interval, req.started}
		gs.timerTab.Insert(tref, req.op)
		gs.intervalTab.Insert(iref, &intervalArgs{tref, req.op.pid})

		return gs.CallReplyTimeout(iref, gs.timeout(sysTime))

	case *timerRef:

		gs.timerTab.Delete(req)

	case Ref:

		if v := gs.intervalTab.Lookup(req); v != nil {
			op := v.(*intervalArgs)
			gs.timerTab.Delete(op.tref)
			gs.intervalTab.Delete(req)
		}
	}

	return gs.CallReplyTimeout("ok", gs.nextTimeout())
}

func (gs *tgs) HandleInfo(req Term) Term {

	switch req := req.(type) {

	case GsTimeout:
		gs.NoReplyTimeout(gs.timeout(time.Now()))

	case *ExitPidReq:
		gs.cancelTimersByPid(req.From)
	}

	return gs.NoReplyTimeout(gs.nextTimeout())
}

// ---------------------------------------------------------------------------
// Locals
// ---------------------------------------------------------------------------

//
// State
//
type tgs struct {
	GenServerSys

	timerTab    Gts
	intervalTab Gts
}

//
// Messages
//
type timerArgs struct {
	pid *Pid
	msg Term
}

type intervalArgs struct {
	tref *timerRef
	pid  *Pid
}

type timerAfterReq struct {
	timeout time.Duration
	op      *timerArgs
	started time.Time
}

type timerIntervalReq struct {
	timeout  time.Duration
	op       *timerArgs
	started  time.Time
	interval uint32
}

//
// Key in timerTab
//
type timerRef struct {
	when     time.Time
	ref      Ref
	interval uint32
	started  time.Time // for debug output
}

func cmpTimerRef(a1, b1 interface{}) int {

	a := a1.(*timerRef)
	b := b1.(*timerRef)

	switch {
	case a.when.After(b.when):
		return 1
	case a.when.Before(b.when):
		return -1
	default:
		return CompareRefs(a.ref, b.ref)

	}
}

//
// Calculate next msg time
//
func (gs *tgs) timeout(sysTime time.Time) time.Duration {

	for k, v, ok := gs.timerTab.First(); ok; k, v, ok = gs.timerTab.First() {

		tref := k.(*timerRef)

		switch {

		case tref.when.After(sysTime):

			return tref.when.Sub(sysTime)

		default:

			gs.timerTab.Delete(tref)

			op := v.(*timerArgs)
			gs.send(op)

			if tref.interval > 0 {
				tref.started = sysTime
				tref.when =
					sysTime.Add(
						time.Duration(tref.interval) * time.Millisecond)
				gs.timerTab.Insert(tref, op)
			}
		}
	}

	return 0
}

func (gs *tgs) nextTimeout() time.Duration {

	k, _, ok := gs.timerTab.First()
	if !ok {
		return time.Duration(0)
	}

	tref := k.(*timerRef)
	timeout := tref.when.Sub(time.Now())
	if timeout.Nanoseconds() > 500 {
		return timeout
	}

	return time.Duration(100) * time.Microsecond
}

func (gs *tgs) cancelTimersByPid(pid *Pid) {

	gs.intervalTab.ForEach(func(k, v interface{}) bool {
		op := v.(*intervalArgs)
		if op.pid.Equal(pid) {
			gs.timerTab.Delete(op.tref)
			return false
		}
		return true
	})
}

func (gs *tgs) send(op *timerArgs) {
	_ = op.pid.Send(op.msg)
}

//
// test
//
func timerServerStop() {
	_ = timerPid.Stop()
	timerPid = nil
}
