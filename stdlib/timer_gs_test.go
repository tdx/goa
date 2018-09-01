package stdlib

import (
	"fmt"
	"testing"
	"time"
)

func TestTimerGsStart(t *testing.T) {
	if err := TimerServerStart(); err != nil {
		t.Fatal(err)
	}
	if err := TimerServerStart(); err != nil {
		t.Fatal(err)
	}

	timerServerStop()

}

func TestTimerGsSend(t *testing.T) {

	if err := TimerServerStart(); err != nil {
		t.Fatal(err)
	}

	pid, err := GenServerStart(new(ts2))
	if err != nil {
		t.Fatal(err)
	}
	defer pid.Stop()

	_, err = pid.Call(&startTimerReq{30, "after 30", false})
	if err != nil {
		t.Fatal(err)
	}

	_, err = pid.Call(&startTimerReq{15, "after 15", false})

	time.Sleep(time.Duration(100) * time.Millisecond)

	timerServerStop()
}

func TestTimerGsSendCancel(t *testing.T) {

	if err := TimerServerStart(); err != nil {
		t.Fatal(err)
	}

	pid, err := GenServerStart(new(ts2))
	if err != nil {
		t.Fatal(err)
	}
	defer pid.Stop()

	tref, err := pid.Call(&startTimerReq{500, "msg1", false})
	if err != nil {
		t.Fatal(err)
	}

	if _, err = pid.Call(&stopTimerReq{tref}); err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Duration(1) * time.Second)

	timerServerStop()
}

func TestTimerGsInterval(t *testing.T) {

	if err := TimerServerStart(); err != nil {
		t.Fatal(err)
	}

	pid, err := GenServerStart(new(ts2))
	if err != nil {
		t.Fatal(err)
	}
	defer pid.Stop()

	_, err = pid.Call(&startTimerReq{30, "msg interval 1", true})
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Duration(400) * time.Millisecond)

	timerServerStop()
}

func TestTimerGsIntervalCancel(t *testing.T) {

	if err := TimerServerStart(); err != nil {
		t.Fatal(err)
	}

	pid, err := GenServerStart(new(ts2))
	if err != nil {
		t.Fatal(err)
	}
	defer pid.Stop()

	tref, err := pid.Call(&startTimerReq{30, "msg interval 2", true})
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Duration(100) * time.Millisecond)

	if _, err = pid.Call(&stopTimerReq{tref}); err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Duration(1) * time.Second)

	timerServerStop()
}

func TestTimerGsIntervalCancelPidExit(t *testing.T) {

	if err := TimerServerStart(); err != nil {
		t.Fatal(err)
	}

	pid, err := GenServerStart(new(ts2))
	if err != nil {
		t.Fatal(err)
	}

	_, err = pid.Call(&startTimerReq{30, "msg interval 2", true})
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Duration(100) * time.Millisecond)

	if err := pid.Stop(); err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Duration(1) * time.Second)

	timerServerStop()
}

//
// GenServer to catch timer events
//
type ts2 struct {
	GenServerSys
}

//
// Messages
//
type startTimerReq struct {
	timeMs   uint32
	msg      string
	interval bool
}

type stopTimerReq struct {
	tref Term
}

type msg struct {
	msg     string
	started time.Time
}

func (gs *ts2) Init(args ...Term) Term {

	if len(args) > 0 {
		for _, arg := range args {
			switch arg := arg.(type) {
			case string:
				switch arg {
				case "debug":
					gs.SetTracer(TraceToConsole())
				}
			}
		}
	}

	return GsInitOk
}

func (gs *ts2) HandleCall(req Term, from From) Term {

	switch req := req.(type) {

	case *startTimerReq:

		var tref Term
		var err error

		m := &msg{req.msg, time.Now()}
		if req.interval {
			tref, err = TimerSendInterval(req.timeMs, gs.Self(), m)
		} else {
			tref, err = TimerSendAfter(req.timeMs, gs.Self(), m)
		}

		if err != nil {
			return err
		}
		return gs.CallReply(tref)

	case *stopTimerReq:

		if err := TimerCancel(req.tref); err != nil {
			return err
		}
	}

	return GsCallReplyOk
}

func (gs *ts2) HandleInfo(req Term) Term {

	switch req := req.(type) {
	case *msg:
		fmt.Println("info:", req.msg, time.Now().Sub(req.started))
	}

	return GsNoReply
}
