package stdlib

import (
	// "fmt"
	"sync/atomic"
	"testing"
	"time"
)

//
// Timers
//
func TestTimerNil(t *testing.T) {

	var timer *Timer
	timer.Stop()
}

func TestTimerGo(t *testing.T) {
	var gotEvent int64

	d := time.Duration(300) * time.Millisecond
	start := time.Now().Truncate(time.Nanosecond)

	sendFunc := func() {

		end := time.Now().Truncate(time.Nanosecond)
		t.Logf("%s timer fired: time %s\n", end, end.Sub(start))

		atomic.AddInt64(&gotEvent, 1)
	}

	t.Logf("%s shedule timer, %s\n", start, d)

	time.AfterFunc(d, sendFunc)

	time.Sleep(time.Duration(1000) * time.Millisecond)

	if atomic.LoadInt64(&gotEvent) < 1 {
		t.Errorf("%s timer not fired: %#v", time.Now(), gotEvent)
	}
}

func TestTimerSendAfterOutsideProcess(t *testing.T) {

	pid, err := GenServerStart(new(ts))
	if err != nil {
		t.Fatal(err)
	}

	m := &testTimerMsg{
		created: time.Now(),
	}

	t.Run("timerGroup", func(t *testing.T) {
		t.Run("timerWorker", func(t *testing.T) {
			t.Parallel()

			f := func() {
				m.fired = time.Now()
				_ = pid.SendInfo(m)
			}

			pid.RunAfter(f, 55)
		})
	})

	time.Sleep(time.Duration(80) * time.Millisecond)

	reply, err := pid.Call("getTimeout")
	if err != nil {
		t.Fatal(err)
	}
	if reply != "true" {
		t.Fatalf("expected reply 'true', actual '%s'", reply)
	}

	if err = pid.Stop(); err != nil {
		t.Fatal(err)
	}

}

func TestTimerStop(t *testing.T) {

	pid, err := GenServerStart(new(ts))
	if err != nil {
		t.Error(err)
	}

	timer := pid.SendAfter("timer", 55)

	time.Sleep(time.Duration(10) * time.Millisecond)

	timer.Stop()

	time.Sleep(time.Duration(50) * time.Millisecond)

	reply, err := pid.Call("getTimeout")
	if err != nil {
		t.Fatal(err)
	}
	if reply != "false" {
		t.Fatalf("expected reply 'false', actual '%s'", reply)
	}

	err = pid.Stop()
	if err != nil {
		t.Fatal(err)
	}
}
