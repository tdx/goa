package stdlib

import (
	"fmt"
	"testing"
	"time"
)

func loop(gp GenProc, args ...Term) error {

	pid := gp.Self()
	sys := pid.GetSysChannel()
	usr := pid.GetUsrChannel()

	lastExitReason := ""

	if len(args) > 0 {
		for _, arg := range args {
			switch arg := arg.(type) {
			case string:
				switch arg {
				case "trapExit":
					gp.SetTrapExit(true)
				case "debug":
					gp.SetTracer(TraceToConsole())
				case "errorInit":
					return fmt.Errorf(arg)
				}
			}
		}
	}

	for {
		select {
		case m := <-sys:
			err := gp.HandleSysMsg(m)
			if err != nil {
				return err
			}

		case m := <-usr:
			switch r := m.(type) {

			case *SyncReq:

				// TraceCall(gp.Tracer(), gp.Self(), "SyncReq:", r.Data)

				switch data := r.Data.(type) {
				case string:
					switch data {
					case "alive":
						r.ReplyChan <- true
						continue
					case "exitReason":
						r.ReplyChan <- lastExitReason
						continue
					}
				}
				r.ReplyChan <- "ok"

			case *SysReq:

				// TraceCall(gp.Tracer(), gp.Self(), "SysReq:", r)

				switch e := r.Data.(type) {
				case *ExitPidReq:
					lastExitReason = e.Reason
				}
			}

		}
	}
}

func TestGenProcTest(t *testing.T) {

	// pid, err := env.spawn(loop, "debug", "trapExit")
	pid, err := env.Spawn(loop)
	if err != nil {
		t.Fatal(err)
	} else if pid == nil {
		t.Fatalf("pid is nil")
	}
	defer pid.Stop()

	if err := pid.Send("hello, beautiful gen proc!"); err != nil {
		t.Fatal(err)
	}

	if reply, err := pid.Call("give me something"); err != nil {
		t.Fatal(err)
	} else if reply != "ok" {
		t.Fatalf("expected ok, actual %#v", reply)
	}

	if err := pid.Exit(ExitNormal); err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Duration(100) * time.Millisecond)
}

func TestGenProcErrorInFunc(t *testing.T) {

	pid, err := env.Spawn(loop, "errorInit")
	if err != nil {
		t.Fatal(err)
	} else if pid == nil {
		t.Fatalf("pid is nil")
	}
	defer pid.Stop()

	if _, err := pid.Call("give me something"); !IsNoProcError(err) {
		t.Fatalf("expected error %s, actual %v", NoProcError, err)
	}
}

func TestGenProcExit1(t *testing.T) {
	//
	// send to yourself
	//
	t1 := testExit{
		arg:              "",
		exitReason:       ExitNormal,
		sleepMs:          50,
		expectedErr:      nil,
		expectedIsAlive:  false,
		expectedLinksLen: 0,
	}

	t2 := t1
	t2.exitReason = "bad"

	t3 := t1
	t3.exitReason = ExitKill

	//
	// with trap exit
	//
	t4 := t1
	t4.arg = "trapExit"

	t5 := t1
	t5.arg = "trapExit"
	t5.exitReason = "bad"

	t6 := t1
	t6.arg = "trapExit"
	t6.exitReason = ExitKill

	tests := []*testExit{
		&t1, &t2, &t3,
		&t4, &t5, &t6,
	}

	for _, s := range tests {
		// fmt.Printf("test: %d\n", i+1)
		exit(s, t)
	}
}

func TestGenProcExit2(t *testing.T) {
	//
	// without trap exit
	//
	t1 := testExit2{
		pid1: testExit{
			arg:              "",
			exitReason:       "",
			expectedErr:      nil,
			expectedIsAlive:  true,
			expectedLinksLen: 1,
		},
		pid2: testExit{
			arg:              "",
			exitReason:       ExitNormal,
			expectedErr:      nil,
			expectedIsAlive:  true,
			expectedLinksLen: 1,
		},
		sleepMs: 50,
	}

	t11 := t1
	t11.pid1.exitReason = ExitNormal
	t11.pid1.expectedLinksLen = 0
	t11.pid2.expectedIsAlive = false

	t2 := t1
	t2.pid1.expectedIsAlive = false
	t2.pid2.exitReason = "bad"
	t2.pid2.expectedIsAlive = false

	t3 := t1
	t3.pid1.expectedIsAlive = false
	t3.pid2.exitReason = ExitKill
	t3.pid2.expectedIsAlive = false

	//
	// with trap exit
	//
	t4 := t1 // ExitNormal
	t4.pid1.arg = "trapExit"
	t4.pid1.expectedLinksLen = 1
	t4.pid2.expectedLinksLen = 1

	t5 := t1
	t5.pid1.arg = "trapExit"
	t5.pid1.expectedLinksLen = 0
	t5.pid1.expectedLastReceivedExit = "bad"
	t5.pid2.exitReason = "bad"
	t5.pid2.expectedIsAlive = false

	t6 := t1
	t6.pid1.arg = "trapExit"
	t6.pid1.expectedLinksLen = 0
	t6.pid1.expectedLastReceivedExit = ExitKilled
	t6.pid2.exitReason = ExitKill
	t6.pid2.expectedIsAlive = false

	t7 := t1
	t7.pid1.arg = "trapExit"
	t7.pid1.expectedIsAlive = true
	t7.pid1.expectedLinksLen = 1
	t7.pid2.arg = "trapExit"
	t7.pid2.exitReason = ExitNormal
	t7.pid2.expectedIsAlive = true
	t7.pid2.expectedLinksLen = 1
	t7.pid2.expectedLastReceivedExit = ExitNormal

	t8 := t7
	t8.pid2.exitReason = "bad"
	t8.pid2.expectedLastReceivedExit = "bad"

	t9 := t7
	t9.pid1.expectedLinksLen = 0
	t9.pid1.expectedLastReceivedExit = ExitKilled
	t9.pid2.exitReason = ExitKill
	t9.pid2.expectedIsAlive = false

	tests := []*testExit2{
		&t1, &t11, &t2, &t3,
		&t4, &t5, &t6,
		&t7, &t8, &t9,
	}
	// tests := []*testExit2{
	// 	&t5,
	// }
	for _, s := range tests {
		// fmt.Printf("test: %d\n", i+1)
		exit2(s, t)
	}

	//                   |  A no trap        |  B no trap
	// A:exit(B, normal) | alive, no msg     | alive, no msg
	// A:exit(B, bad)    | stopped, no msg   | stopped, no msg
	// A:exit(B, kill)   | stopped, killed   | stopped, killed
	//
	// stop A normal     | stopped,          | alive, no msg, links[]
	//                   | Terminate(normal) |
	// stop A bad        | stopped,          | stopped, no msg
	//                   | Terminate(bad)    |

	//                   |  A    trap        |  B no trap
	// A:exit(B, normal) | alive, no msg     | alive, no msg
	// A:exit(B, bad)    | stopped,          | stopped, no msg
	//                     {exit,Pid2,bad}   |
	// stop A normal     | stopped,          | alive, no msg, links[]
	//                   | Terminate(normal) |
	// stop A bad        | stopped,          | stopped, no msg
	//                   | Terminate(bad)    |
	// stop B normal     | alive,            | stopped, Terminate(normal)
	//                   | {exit,Pid2,normal}|
	// stop B bad        | alive,            | stopped, Terminate(bad)
	//                   | {exit,Pid2,bad}   |

	//                   |  A    trap                |  B trap
	// A:exit(B, normal) | alive, no msg             | alive, {exit,Pid1,normal}
	// A:exit(B, bad)    | alive, no msg             | alive, {exit,Pid1,bad}
	// A:exit(B, kill)   | alive, {exit,Pid2,killed} | stopped, no msg
}

//
// Locals
//
type testExit2 struct {
	pid1    testExit
	pid2    testExit
	sleepMs int
}

type testExit struct {
	arg                      string
	exitReason               string
	sleepMs                  int
	expectedErr              error
	expectedIsAlive          bool
	expectedLinksLen         int
	expectedLastReceivedExit string
}

func start(t *testing.T, args ...Term) *Pid {
	pid, err := env.Spawn(loop, args...)
	if err != nil {
		t.Fatal(err)
	} else if pid == nil {
		t.Fatalf("pid is nil")
	}
	return pid
}

func startLink(t *testing.T, pidFrom *Pid, args ...Term) *Pid {
	pid, err := pidFrom.SpawnLink(loop, args...)
	if err != nil {
		t.Fatal(err)
	} else if pid == nil {
		t.Fatalf("pid is nil")
	}
	return pid
}

func isAlive(pid *Pid, t *testing.T) bool {
	alive, err := pid.Call("alive")
	if err != nil {
		if !IsNoProcError(err) {
			t.Log(err)
		}
		return false
	}

	switch alive := alive.(type) {
	case bool:
		return alive
	default:
		t.Fatalf("isAlive: unexpected return type: %#v", alive)
	}
	return false
}

func exitReason(pid *Pid, t *testing.T) string {
	lastExit, err := pid.Call("exitReason")
	if err != nil {
		t.Error(err)
		return "call failed"
	}

	switch lastExit := lastExit.(type) {
	case string:
		return lastExit
	}
	return "bad call value"
}

func exit(test *testExit, t *testing.T) {
	pid := start(t, test.arg)

	err := pid.Exit(test.exitReason)
	if err != test.expectedErr {
		t.Fatalf("expected %s error, actual %s", test.expectedErr, err)
		return
	}

	if test.sleepMs > 0 {
		time.Sleep(time.Duration(test.sleepMs) * time.Millisecond)
	}

	alive := isAlive(pid, t)
	if test.expectedIsAlive != alive {
		t.Fatalf("expected alive %v, actual %v",
			test.expectedIsAlive, alive)
		return
	}

	if test.expectedIsAlive && test.expectedLinksLen >= 0 {
		if links, err := pid.ProcessLinks(); err != nil {
			t.Fatal(err)
			return
		} else if len(links) != test.expectedLinksLen {
			t.Fatalf("expected links %d, actual %d",
				test.expectedLinksLen, len(links))
			return
		}
	}

	if test.expectedIsAlive {
		lastExit := exitReason(pid, t)
		if test.expectedLastReceivedExit != lastExit {
			t.Fatalf("expected last received exit: %s, actual %s",
				test.expectedLastReceivedExit, lastExit)
			return
		}
	}
}

func exit2(test *testExit2, t *testing.T) {
	pid1 := start(t, test.pid1.arg)
	pid2 := startLink(t, pid1, test.pid2.arg)

	var err error

	if test.pid1.exitReason == test.pid2.exitReason {
		err = pid2.Exit(test.pid2.exitReason) // send exit to self
	} else {
		err = pid1.ExitReason(pid2, test.pid2.exitReason)
	}

	if err != test.pid2.expectedErr {
		t.Fatalf("expected %s error, actual %s", test.pid2.expectedErr, err)
		return
	}

	if test.sleepMs > 0 {
		time.Sleep(time.Duration(test.sleepMs) * time.Millisecond)
	}

	alive := isAlive(pid1, t)
	if test.pid1.expectedIsAlive != alive {
		t.Fatalf("expected pid1 alive %v, actual %v",
			test.pid1.expectedIsAlive, alive)
		return
	}

	alive = isAlive(pid2, t)
	if test.pid2.expectedIsAlive != alive {
		t.Fatalf("expected pid2 alive %v, actual %v",
			test.pid2.expectedIsAlive, alive)
		return
	}

	if test.pid1.expectedIsAlive && test.pid1.expectedLinksLen >= 0 {
		if links1, err := pid1.ProcessLinks(); err != nil {
			t.Fatal(err)
			return
		} else if len(links1) != test.pid1.expectedLinksLen {
			t.Fatalf("expected pi1 links %d, actual %d",
				test.pid1.expectedLinksLen, len(links1))
			return
		}
	}

	if test.pid2.expectedIsAlive && test.pid2.expectedLinksLen >= 0 {
		if links2, err := pid2.ProcessLinks(); err != nil {
			t.Fatal(err)
			return
		} else if len(links2) != test.pid2.expectedLinksLen {
			t.Fatalf("expected pid2 links %d, actual %d",
				test.pid2.expectedLinksLen, len(links2))
			return
		}
	}

	if test.pid1.expectedIsAlive {
		lastExit := exitReason(pid1, t)
		if test.pid1.expectedLastReceivedExit != lastExit {
			t.Fatalf("expected pid1 last received exit: %s, actual %s",
				test.pid1.expectedLastReceivedExit, lastExit)
			return
		}
	}

	if test.pid2.expectedIsAlive {
		lastExit := exitReason(pid2, t)
		if test.pid2.expectedLastReceivedExit != lastExit {
			t.Fatalf("expected pid2 last received exit: %s, actual %s",
				test.pid2.expectedLastReceivedExit, lastExit)
			return
		}
	}
}
