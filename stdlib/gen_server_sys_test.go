package stdlib

import (
	"errors"
	"fmt"
	"runtime"
	"testing"
	"time"
)

func TestGenServerSysCreate(t *testing.T) {

	_, err := GenServerStart(nil)
	if err == nil {
		t.Fatal("expected 'GenServer parameter is nil', actual no error")
	}

	pid, err := GenServerStart(new(GenServerSys))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := pid.Call("alive"); err != nil {
		t.Fatal(err)
	}

	if err := pid.Stop(); err != nil {
		t.Fatal(err)
	}

	if err := pid.Cast("alive"); !IsNoProcError(err) {
		t.Fatalf("%s: expected '%s' error, actual '%v'", pid, NoProcError, err)
	}
}

func TestGenServerStopPidInDiffRoutines(t *testing.T) {
	pid, err := GenServerStart(new(GenServerSys))
	if err != nil {
		t.Fatal(err)
	}

	t.Run("stop", func(t *testing.T) {
		t.Run("process1", func(t *testing.T) {
			t.Parallel()
			err := pid.Stop()
			if err != nil && !IsNoProcError(err) {
				t.Fatalf("expected '%s' error, actual '%s'", NoProcError, err)
			}
		})
	})

	err = pid.Stop()
	if err != nil && !IsNoProcError(err) {
		t.Fatalf("expected '%s' error, actual '%s'", NoProcError, err)
	}
}

func TestGenServerInitOk(t *testing.T) {

	pid, err := GenServerStart(new(ts))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := pid.Call("alive"); err != nil {
		t.Fatal(err)
	}

	if err := pid.Stop(); err != nil {
		t.Fatal(err)
	}

	if err := pid.Cast("alive"); !IsNoProcError(err) {
		t.Fatalf("expected '%s' error, actual %v", NoProcError, err)
	}
}

func TestGenServerInitError(t *testing.T) {

	_, err := GenServerStart(new(ts), "crash")
	if err == nil {
		t.Fatal("expected error, actual no error")
	}

	_, err = GenServerStart(new(ts), "badReturn")
	if err == nil {
		t.Fatal("expected error, actual no error")
	}

	_, err = GenServerStart(new(ts), "errorInit")
	if err == nil {
		t.Fatalf("must fail in Init")
	}

	_, err = GenServerStart(new(ts), "initStopExitNormal")
	if !IsExitNormalError(err) {
		t.Fatalf("expected '%s' error, actual %s", ExitNormal, err)
	}

	_, err = GenServerStart(new(ts), "initStopBad")
	if err == nil {
		t.Fatal("expected error, actual no error")
	}

	pid, err := GenServerStart(new(ts), "initTimeout")
	if err != nil {
		t.Fatal(err)
	}
	defer pid.Stop()

	time.Sleep(time.Duration(50) * time.Millisecond)

	reply, err := pid.Call("getTimeout")
	if err != nil {
		t.Fatal(err)
	}
	if reply != "true" {
		t.Fatalf("expected reply 'true', actual '%s'", reply)
	}
}

func TestGenServerCall(t *testing.T) {
	t1 := testGs{
		initArg:         "",
		op:              "call",
		opArg:           "ping",
		expectedOpReply: "pong",
		expectedOpErr:   "",
		expectedOpAlive: true,
		exitReason:      ExitNormal,
		expectedExitErr: nil,
		expectedIsAlive: false,
		sleepMs:         100,
	}

	t2 := t1
	t2.initArg = ""
	t2.opArg = "error"
	t2.expectedOpErr = "Call return error"
	t2.expectedOpAlive = false

	t3 := t1
	t3.opArg = "stop"
	t3.expectedOpReply = "ok"
	t3.expectedOpAlive = false

	t4 := t1
	t4.opArg = "crash"
	t4.expectedOpErr = "runtime error: integer divide by zero"
	t4.expectedOpAlive = false

	t5 := t1
	t5.opArg = "any"
	t5.expectedOpErr = `HandleCall bad reply: "any"`
	t5.expectedOpAlive = false

	//
	// timeouts
	//  TODO: redesign, sometimes timeouts arrive much later
	//
	t6 := t1
	// t6.initArg = "debug"
	t6.opArg = &timeoutReq{"exitTimeout", 50, true, ExitNormal}
	t6.expectedOpReply = "ok"
	t6.expectedOpExitTimeout = 150

	t7 := t1
	t7.opArg = &timeoutReq{"setTimeout", 50, false, ExitNormal}
	t7.expectedOpReply = "ok"
	t7.exitReason = "savePid"

	t8 := t7
	t8.opArg = "getTimeout"
	t8.expectedOpReply = "true"
	t8.exitReason = ExitNormal

	t9 := t1
	t9.opArg = &timeoutReq{"setTimeout", 50, true, ExitNormal}
	t9.expectedOpReply = "ok"
	t9.expectedOpExitTimeout = 150

	t10 := t1
	t10.opArg = &timeoutReq{"setTimeout", 50, true, "return error"}
	t10.expectedOpReply = "ok"
	t10.expectedOpExitTimeout = 150

	t11 := t1
	t11.opArg = &timeoutReq{"setTimeout", 50, true, "any"}
	t11.expectedOpReply = "ok"
	t11.expectedOpExitTimeout = 150

	tests := []*testGs{
		&t1, &t2, &t3, &t4, &t5,
		&t6, &t7, &t8, &t9, &t10, &t11,
	}
	var pid *Pid
	for i := range tests {
		s := tests[i]
		t.Run(fmt.Sprintf("%v:", i+1),
			func(t *testing.T) {
				pid = doTestGs(i, pid, s, t)
			})
	}
	// fmt.Printf("\n")
}

func TestGenServerCallNoReply(t *testing.T) {
	pid, err := GenServerStart(new(ts))
	if err != nil {
		t.Fatal(err)
	}
	defer pid.Stop()

	reply, err := pid.Call("noReply")
	if err != nil {
		t.Fatal(err)
	}

	if reply != true {
		t.Fatalf("exprected boolean return 'true', actual %#v", reply)
	}

	_, err = pid.Call("noReplyTimeout")
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Duration(60) * time.Millisecond)

	reply, err = pid.Call("getTimeout")
	if err != nil {
		t.Fatal(err)
	}
	if reply != "true" {
		t.Fatalf("expected reply 'true', actual '%s'", reply)
	}
}

func TestGenServerSend(t *testing.T) {

	pid, err := GenServerStart(new(ts))
	if err != nil {
		t.Fatal(err)
	}

	if err = pid.Cast("return error"); err != nil {
		t.Fatal(err)
	}
	_, err = pid.Call("ping")
	if err != nil && !IsNoProcError(err) {
		t.Fatalf("expected '%s' error, actual '%s'", NoProcError, err)
	}

	pid, err = GenServerStart(new(ts))
	if err != nil {
		t.Fatal(err)
	}

	if err = pid.Cast("stopNormal"); err != nil {
		t.Fatal(err)
	}
	_, err = pid.Call("ping")
	if !IsNoProcError(err) {
		t.Fatalf("expected '%s' error, actual '%s'", NoProcError, err)
	}

	pid, err = GenServerStart(new(ts))
	if err != nil {
		t.Fatal(err)
	}

	if err = pid.Cast("stopBad"); err != nil {
		t.Fatal(err)
	}
	_, err = pid.Call("ping")
	if !IsNoProcError(err) {
		t.Fatalf("expected '%s' error, actual '%s'", NoProcError, err)
	}

	pid, err = GenServerStart(new(ts))
	if err != nil {
		t.Fatal(err)
	}
	defer pid.Stop()

	if err = pid.Cast("timeout"); err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Duration(150) * time.Millisecond)

	reply, err := pid.Call("getTimeout")
	if err != nil {
		t.Fatal(err)
	}
	if reply != "true" {
		t.Fatalf("expected reply 'true', actual '%s'", reply)
	}
}

func TestGenServerLink(t *testing.T) {
	pid, err := GenServerStart(new(ts), "trapExit")
	if err != nil {
		t.Fatal(err)
	}
	defer pid.Stop()

	reply, err := pid.Call("spawnLink")
	if err != nil {
		t.Fatal(err)
	}
	newPid := reply.(*Pid)
	defer newPid.Stop()

	links, err := newPid.ProcessLinks()
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 || links[0] != pid {
		t.Fatalf("wrong links in process: %v\n", links)
	}

	err = newPid.Stop()
	if err != nil {
		t.Fatal(err)
	}

	links, err = pid.ProcessLinks()
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 0 {
		t.Fatalf("expected process has no links, actual %v\n", links)
	}
}

func TestGenServerUnlinkOk(t *testing.T) {
	pid, err := GenServerStart(new(ts), "trapExit")
	if err != nil {
		t.Fatal(err)
	}
	defer pid.Stop()

	reply, err := pid.Call("testUnlink")
	if err != nil {
		t.Fatal(err)
	}

	var funcPid *Pid
	switch fpid := reply.(type) {
	case *Pid:
		funcPid = fpid
		time.Sleep(time.Duration(30) * time.Millisecond)
	default:
		t.Fatalf("exporecte reply '*Pid', actual '%#v'", funcPid)
	}

	links, err := pid.ProcessLinks()
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 || links[0] != funcPid {
		t.Fatalf("wrong links in process: %v\n", links)
	}

	if _, err = pid.Call(&unlinkReq{funcPid}); err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Duration(30) * time.Millisecond)

	links, err = pid.ProcessLinks()
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 0 {
		t.Fatalf("expected process has no links, actual %v\n", links)
	}
}

func TestGenServerUnlinkFromFastExitGenProc(t *testing.T) {
	pid, err := GenServerStart(new(ts), "trapExit")
	if err != nil {
		t.Fatal(err)
	}
	defer pid.Stop()

	_, err = pid.Call("testUnlinkFastExit")
	if err != nil {
		t.Fatal(err)
	}

	var i int
	var links []*Pid
	delay := time.Duration(20) * time.Millisecond

	for i = 0; i < 5; i++ {

		links, err = pid.ProcessLinks()
		if err != nil {
			t.Fatal(err)
		}
		if len(links) != 0 {
			time.Sleep(delay)
		} else {
			break
		}

	}
	if len(links) != 0 {
		t.Fatalf("after %d ms wrong links in process: %v\n",
			i*5, links)
	}
}

func TestGenServerMonitorDownAfterMonitorOk(t *testing.T) {
	pid, err := GenServerStart(new(ts))
	if err != nil {
		t.Fatal(err)
	}
	defer pid.Stop()

	reply, err := pid.Call("startMonitor")
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Duration(30) * time.Millisecond)

	switch funcPid := reply.(type) {
	case *Pid:
		if err := funcPid.Send("exit"); err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Duration(30) * time.Millisecond)
	default:
		t.Fatalf("expected reply '*Pid', actual '%#v'", funcPid)
	}

	reply, err = pid.Call("monitorMessage")
	if err != nil {
		t.Fatal(err)
	}
	if reply != "exit" {
		t.Fatalf("expected reply 'exit', actual '%s'", reply)
	}
}

// func TestGenServerMonitorDownInMonitor(t *testing.T) {
// 	pid, err := GenServerStart(new(ts))
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer pid.Stop()

// 	_, err = pid.Call("startBadMonitor")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	time.Sleep(time.Duration(30) * time.Millisecond)

// 	reply, err := pid.Call("monitorMessage")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if reply != NoProc {
// 		t.Fatalf("expected reply '%s', actual '%s'", NoProc, reply)
// 	}
// }

//
// Expected MonitorProcessPid in handler of startBadMonitor2 does not found
//  running process
//
// func TestGenServerMonitorDownFromFlush(t *testing.T) {
// 	pid, err := GenServerStart(new(ts))
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer pid.Stop()

// 	_, err = pid.Call("startBadMonitor2")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	reply, err := pid.Call("monitorMessage")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if reply != NoProc {
// 		t.Fatalf("expected reply '%s', actual '%s'", NoProc, reply)
// 	}
// }

func TestGenServerMonitorDemonitorNoMessage(t *testing.T) {
	pid, err := GenServerStart(new(ts))
	if err != nil {
		t.Fatal(err)
	}
	defer pid.Stop()

	reply, err := pid.Call("demonitorOk")
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Duration(30) * time.Millisecond)

	switch funcPid := reply.(type) {
	case *Pid:
		if err := funcPid.Cast("exit"); err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Duration(30) * time.Millisecond)
	default:
		t.Fatalf("exporecte reply '*Pid', actual '%#v'", funcPid)
	}

	reply, err = pid.Call("monitorMessage")
	if err != nil {
		t.Fatal(err)
	}
	if reply != "" {
		t.Fatalf("expected reply '', actual '%s'", reply)
	}
}

func TestGenServerStartOpts(t *testing.T) {

	opts := NewSpawnOpts().WithName("test1")
	pid, err := GenServerStartOpts(new(ts), opts)
	if err != nil {
		t.Fatal(err)
	}
	defer pid.Stop()

	opts2 := NewSpawnOpts().WithName("test1").WithSpawnOrLocate()
	pid2, err := GenServerStartOpts(new(ts), opts2)
	if err != nil {
		t.Fatal(err)
	}
	defer pid2.Stop()

	if pid.ID() != pid2.ID() {
		t.Fatalf("expected pid2=%s, actual %s", pid, pid2)
	}

	//
	// must fail - AlreadyRegError
	//
	opts3 := NewSpawnOpts().WithName("test1")
	_, err = GenServerStartOpts(new(ts), opts3)
	if !IsAlreadyRegError(err) {
		t.Fatalf("exptected '%s' error, actual '%s", AlreadyRegError, err)
	}

	//
	// with prefix
	//
	opts4 := NewSpawnOpts().WithName("test1").WithPrefix("group1")
	pid4, err := GenServerStartOpts(new(ts), opts4)
	if err != nil {
		t.Fatal(err)
	}
	defer pid4.Stop()

	opts5 := NewSpawnOpts().
		WithName("test1").WithPrefix("group1").WithSpawnOrLocate()
	pid5, err := GenServerStartOpts(new(ts), opts5, "debug")
	if err != nil {
		t.Fatal(err)
	}
	defer pid5.Stop()

	if pid4.ID() != pid5.ID() {
		t.Fatalf("expected pid2=%s, actual %s", pid4, pid5)
	}

	//
	// must fail - AlreadyRegError
	//
	opts3 = NewSpawnOpts().WithName("test1").WithPrefix("group1")
	_, err = GenServerStartOpts(new(ts), opts3)
	if !IsAlreadyRegError(err) {
		t.Fatalf("exptected '%s' error, actual '%s", AlreadyRegError, err)
	}
}

//
// Locals
//
type testGs struct {
	initArg               string
	op                    string // "call"
	opArg                 Term
	expectedOpReply       string
	expectedOpExitTimeout uint32
	expectedOpErr         string
	expectedOpAlive       bool
	exitReason            string
	expectedExitErr       error
	expectedIsAlive       bool
	sleepMs               int
}

func doTestGs(i int, pid *Pid, s *testGs, t *testing.T) *Pid {

	i++

	var err error

	if pid == nil {
		pid, err = GenServerStart(new(ts), s.initArg)
		if err != nil {
			t.Fatal(err)
			return nil
		}
	}
	if s.exitReason != "savePid" {
		defer pid.Stop()
	}

	switch s.op {
	case "call":
		errStr := ""
		reply, err := pid.Call(s.opArg)
		if err != nil {
			errStr = err.Error()
		}
		if s.expectedOpErr != errStr {
			t.Fatalf("%d: expected '%s' error '%#v', actual '%#v'",
				i, s.op, s.expectedOpErr, errStr)
			return nil
		}

		if err == nil && reply != s.expectedOpReply {
			t.Fatalf("%d: expected '%s' reply '%s', actual '%s'",
				i, s.op, s.expectedOpReply, reply)
			return nil
		}

		if s.expectedOpExitTimeout > 0 {
			//
			// any call to process will remove inactivity timer, so wait + 30 ms
			//
			waitMs := s.expectedOpExitTimeout + 30
			time.Sleep(time.Duration(waitMs) * time.Millisecond)

			alive := isAlive(pid, t)
			if alive {
				t.Fatalf(
					"%d: expected process is died, actual is alive after %d ms\n",
					i, waitMs)
				return nil
			}

			return nil
		}

	default:
		t.Fatalf("%d: unexpected op '%s'", i, s.op)
		return nil
	}

	if s.exitReason == "savePid" {
		if s.sleepMs > 0 {
			time.Sleep(time.Duration(s.sleepMs) * time.Millisecond)
		}

		return pid
	}

	alive := isAlive(pid, t)
	if s.expectedOpAlive != alive {
		t.Fatalf("%d: expected op alive %v, actual %v",
			i, s.expectedOpAlive, alive)
		return nil
	}
	if !alive {
		return nil
	}

	// stop gs
	if err := pid.Exit(s.exitReason); err != s.expectedExitErr {
		t.Fatalf("%d: expected exit error %s, actual %s",
			i, s.expectedExitErr, err)
		return nil
	}

	if s.sleepMs > 0 {
		time.Sleep(time.Duration(s.sleepMs) * time.Millisecond)
	}

	alive = isAlive(pid, t)
	if s.expectedIsAlive != alive {
		t.Fatalf("%d: expected alive %v, actual %v",
			i, s.expectedIsAlive, alive)
		return nil
	}

	return nil
}

//
// Test gen_server
//
type ts struct {
	GenServerSys

	//
	// timeout test
	//
	gotTimeout       bool
	exitAfterTimeout bool
	exitReason       string
	timerStart       time.Time
	//
	// monitor test
	//
	monitorMessage string
}

type timeoutReq struct {
	timeutOp         string
	timeoutMs        uint32
	exitAfterTimeout bool
	exitReason       string
}

type unlinkReq struct {
	pid *Pid
}

type testTimerMsg struct {
	created  time.Time
	fired    time.Time
	accepted time.Time
}

func testFunc(gp GenProc, args ...Term) error {
	TraceCall(gp.Tracer(), gp.Self(), "testFunc", "new gen proc")
	time.Sleep(time.Duration(50) * time.Millisecond)
	return nil
}

func testMonFunc(gp GenProc, args ...Term) error {

	pid := gp.Self()

	TraceCall(gp.Tracer(), gp.Self(), "testMonFunc",
		"new gen proc with sys messages handler")

	sys := pid.GetSysChannel()
	usr := pid.GetUsrChannel()

	for {
		select {
		case m := <-sys:
			if err := gp.HandleSysMsg(m); err != nil {
				return err
			}
		case m := <-usr:
			switch m := m.(type) {
			case string:
				switch m {
				case "exit":
					return errors.New("exit")
				}
				// 	}
			}
		}
	}
}

func testFastExitFunc(gp GenProc, args ...Term) error {
	// gp.SetTracer(TraceToConsole())
	TraceCall(gp.Tracer(), gp.Self(), "testFastExitFunc", "fast exit func")
	return nil
}

func (gs *ts) Init(args ...Term) Term {

	if len(args) > 0 {
		for _, arg := range args {
			switch arg := arg.(type) {
			case string:
				switch arg {
				case "crash":
					a := 10
					a = a / (a - 10)
					fmt.Println(a) // just to hide ineffAsign warning

				case "badReturn":
					return 123

				case "errorInit":
					return fmt.Errorf(arg)

				case "initStopExitNormal":
					return gs.Stop(ExitNormal)

				case "initStopBad":
					return gs.Stop("bad")

				case "initTimeout":
					return gs.InitTimeout(time.Duration(30) * time.Millisecond)

				case "trapExit":
					gs.SetTrapExit(true)

				case "debug":
					gs.SetTracer(TraceToConsole())
				}
			}
		}
	}

	return gs.InitOk()
}

func (gs *ts) HandleCall(req Term, from From) Term {

	switch req := req.(type) {
	case string:

		return gs.handleString(req, from)

	case *timeoutReq:
		switch req.timeutOp {

		case "setTimeout":
			gs.exitAfterTimeout = req.exitAfterTimeout
			gs.exitReason = req.exitReason
			return gs.CallReplyTimeout(
				"ok", time.Duration(req.timeoutMs)*time.Millisecond)

		case "exitTimeout":
			gs.exitAfterTimeout = req.exitAfterTimeout
			gs.exitReason = req.exitReason
			gs.timerStart = time.Now()
			return gs.CallReplyTimeout(
				"ok", time.Duration(req.timeoutMs)*time.Millisecond)
		}

	case *unlinkReq:
		gs.Unlink(req.pid)
		return gs.CallReplyOk()

	}

	return gs.CallReplyOk()
}

func (gs *ts) HandleCast(req Term) Term {

	switch req := req.(type) {
	case string:
		switch req {
		case "return error":
			return errors.New("return error")
		case "stopNormal":
			return gs.Stop(ExitNormal)
		case "stopBad":
			return gs.Stop("bad")
		case "timeout":
			return gs.NoReplyTimeout(20)
		}
	}

	return gs.NoReply()
}

func (gs *ts) HandleInfo(req Term) Term {

	switch req := req.(type) {

	case GsTimeout:
		gs.gotTimeout = true
		timeout := time.Now().Sub(gs.timerStart)

		TraceCall(gs.Tracer(), gs.Self(), "HandleInfo: got timeout", timeout)

		if gs.exitAfterTimeout {
			switch gs.exitReason {
			case "return error":
				return errors.New(gs.exitReason)
			case "any":
				return gs.exitReason
			}
			return gs.Stop(gs.exitReason)
		}

	case *MonitorDownReq:
		gs.monitorMessage = req.Reason

	case string:
		switch req {
		case "timer":
			gs.gotTimeout = true
		}

	case *testTimerMsg:
		now := time.Now()
		gs.gotTimeout = true
		req.accepted = now
		elapsed := now.Sub(req.created)
		if elapsed > (time.Duration(80) * time.Millisecond) {
			fmt.Printf("%s %s: to timer event: %s/%s/%s\n",
				now.Truncate(time.Microsecond), gs.Self(),
				req.fired.Sub(req.created), now.Sub(req.fired), elapsed)
		}
	}

	return gs.NoReply()
}

func (gs *ts) handleString(req string, from From) Term {
	switch req {
	case "ping":
		return gs.CallReply("pong")

	case "alive":
		return gs.CallReply(true)

	case "error":
		return fmt.Errorf("Call return error")

	case "stop":
		return gs.CallStop("stop", "ok")

	case "crash":
		a := 10
		return gs.CallReply(a / (a - 10))

	case "noReply":
		go func() {
			time.Sleep(time.Duration(20) * time.Millisecond)
			gs.Reply(from, true)
		}()
		return gs.NoReply()

	case "noReplyTimeout":
		go func() {
			time.Sleep(time.Duration(20) * time.Millisecond)
			gs.Reply(from, true)
		}()
		return gs.NoReplyTimeout(30)

	case "getTimeout":
		return gs.CallReply(fmt.Sprintf("%v", gs.gotTimeout))

	case "spawnLink":
		pid, err := gs.Self().SpawnLink(testMonFunc)
		if err != nil {
			return err
		}
		return gs.CallReply(pid)

	case "testUnlink":
		pid, err := gs.Self().SpawnLink(testMonFunc)
		if err != nil {
			return err
		}
		return gs.CallReply(pid)

	default:
		return gs.handleString2(req, from)
	}
}

//
// split one big func for gocyclo -avg -over 15
//
func (gs *ts) handleString2(req string, from From) Term {

	switch req {
	case "testUnlinkFastExit":
		pid, err := gs.Self().SpawnLink(testFastExitFunc)
		if err != nil {
			return err
		}
		return gs.CallReply(pid)

	case "startMonitor":
		pid, err := Spawn(testMonFunc)
		if err != nil {
			return err
		}
		gs.MonitorProcessPid(pid)
		return gs.CallReply(pid)

	case "startBadMonitor":
		pid, err := Spawn(testFastExitFunc)
		if err != nil {
			return err
		}
		// wait testFastExitFunc exit
		time.Sleep(time.Duration(20) * time.Millisecond)
		gs.MonitorProcessPid(pid)
		return gs.CallReply(pid)

	case "startBadMonitor2":
		pid, err := Spawn(testFastExitFunc)
		if err != nil {
			return err
		}
		runtime.Gosched()

		//
		// wait for process to exit
		//
		time.Sleep(time.Duration(30) * time.Millisecond)

		gs.MonitorProcessPid(pid) // can be flushed in GenProc
		return gs.CallReply(pid)

	case "demonitorOk":
		pid, err := Spawn(testMonFunc)
		if err != nil {
			return err
		}
		ref := gs.MonitorProcessPid(pid)
		gs.DemonitorProcessPid(ref)
		return gs.CallReply(pid)

	case "monitorMessage":
		return gs.CallReply(gs.monitorMessage)

	default:
		return req // should crash with unexpected return value
	}
}

//
// Bench
//
func BenchmarkCastPid(b *testing.B) {
	pid, err := GenServerStart(new(ts))
	if err == nil {

		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = pid.Cast("ping")
			}
		})
	}
}

func BenchmarkSendPid(b *testing.B) {
	pid, err := GenServerStart(new(ts))
	if err == nil {

		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = pid.Send("ping")
			}
		})
	}
}

func BenchmarkCallPidParallel(b *testing.B) {
	pid, err := GenServerStart(new(ts))
	if err == nil {

		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = pid.Call("ping")
			}
		})
	}
}

func BenchmarkCallPidSync(b *testing.B) {
	pid, err := GenServerStart(new(ts))
	if err == nil {

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _ = pid.Call("ping")
		}
	}
}

func BenchmarkGenServerStart(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, err := GenServerStart(new(ts)); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenServerCallAfterStop(b *testing.B) {
	for i := 0; i < b.N; i++ {
		pid, err := GenServerStart(new(ts))
		if err != nil {
			b.Fatal(i, pid, err)
		}
		if err = pid.Stop(); err != nil {
			b.Fatal(i, pid, err)
		}

		alive, err := pid.Call("alive")

		if err != nil && !IsNoProcError(err) {
			b.Fatalf("expected '%s' error, actual '%s'", NoProcError, err)
		}
		if err == nil && alive != false {
			b.Fatalf("expected alive is false, actual '%#v'", alive)
		}
	}
}
