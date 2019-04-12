package stdlib

import (
	"errors"
	"fmt"
	"runtime"
	"time"
)

const (
	traceFuncDoInit    = "Init"
	traceFuncDoCall    = "HandleCall"
	traceFuncDoCast    = "HandleCast"
	traceFuncDoInfo    = "HandleInfo"
	traceFuncTerminate = "Terminate"
)

//
// GenServerSys is a default implementation of GenServer interface
//
type GenServerSys struct {
	GenProcSys

	initChan   chan error
	callbackGs GenServer
	//
	reply   Term
	timeout time.Duration
	reason  string
}

//
// InitPrepare makes channel for InitAck
//
func (gs *GenServerSys) InitPrepare() {

	gs.initChan = make(chan error, 1)
}

//
// InitAck waits for result of process initialization
//
func (gs *GenServerSys) InitAck() error {

	initResult := <-gs.initChan

	return initResult
}

//
// Init initializes process state using arbitrary arguments
//
func (gs *GenServerSys) Init(args ...Term) Term {
	return gsInitOk
}

//
// HandleCall handles incoming messages from `pid.Call(data)`
//
func (gs *GenServerSys) HandleCall(req Term, from From) Term {
	return gsCallReplyOk
}

//
// HandleCast handles incoming messages from `pid.Cast(data)`
//
func (gs *GenServerSys) HandleCast(req Term) Term {
	return gsNoReply
}

//
// HandleInfo handles timeouts and system messages
//
func (gs *GenServerSys) HandleInfo(req Term) Term {
	return gsNoReply
}

//
// Terminate called when process died
//
func (gs *GenServerSys) Terminate(reason string) {
}

//
// Reply directly to caller
//
func (gs *GenServerSys) Reply(from From, data Term) {
	from <- data
}

func (gs *GenServerSys) setCallback(gp GenServer) {
	gs.callbackGs = gp
}

//
// Callbacks returns
//

//
// InitOk makes gsInitOk reply
//
func (gs *GenServerSys) InitOk() Term {

	return gsInitOk
}

//
// InitTimeout makes InitTimeout reply
//
func (gs *GenServerSys) InitTimeout(timeout time.Duration) Term {

	gs.timeout = timeout

	return gsInitTimeout
}

//
// CallReply makes CallReply reply
//
func (gs *GenServerSys) CallReply(reply Term) Term {

	gs.reply = reply

	return gsCallReply
}

//
// CallReplyOk makes gsCallReplyOk reply
//
func (gs *GenServerSys) CallReplyOk() Term {

	return gsCallReplyOk
}

//
// CallReplyTimeout makes CallReplyTimeout reply
//
func (gs *GenServerSys) CallReplyTimeout(
	reply Term, timeout time.Duration) Term {

	gs.reply = reply
	gs.timeout = timeout

	return gsCallReplyTimeout
}

//
// CallStop makes CallStop reply
//
func (gs *GenServerSys) CallStop(reason string, reply Term) Term {

	gs.reason = reason
	gs.reply = reply

	return gsCallStop
}

//
// Stop makes Stop reply
//
func (gs *GenServerSys) Stop(reason string) Term {

	gs.reason = reason

	return gsStop
}

//
// NoReply makes gsNoReply reply
//
func (gs *GenServerSys) NoReply() Term {

	return gsNoReply
}

//
// NoReplyTimeout makes NoReplyTimeout reply
//
func (gs *GenServerSys) NoReplyTimeout(timeout time.Duration) Term {

	gs.timeout = timeout

	return gsNoReplyTimeout
}

//
// GenProcLoop is a main gen_server loop function
//
func (gs *GenServerSys) GenProcLoop(args ...Term) (err error) {

	pid := gs.Self()

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)

			trace := make([]byte, 4096)
			n := runtime.Stack(trace, false)

			fmt.Println(time.Now().Truncate(time.Microsecond), gs.Self(),
				"crashed with reason:", r, n, "bytes stack:",
				string(trace[:n]))

			TraceCall(gs.Tracer(), gs.Self(), "GenServerSysLoop crashed", err)
		}

		gs.doTerminate(err.Error())
	}()

	sys := pid.GetSysChannel()
	usr := pid.GetUsrChannel()

	var timeout <-chan time.Time

	if timeout, err = gs.doInit(args...); err != nil {
		return
	}

	for {

		//
		// check sys messages first
		//
		select {
		case m := <-sys:
			if err = gs.HandleSysMsg(m); err != nil {
				return
			}
		default:
		}

		select {

		case m := <-sys:
			if err = gs.HandleSysMsg(m); err != nil {
				return
			}

		case m := <-usr:

			switch m := m.(type) {

			case *SyncReq:

				if timeout, err = gs.doCall(m.Data, m.ReplyChan); err != nil {
					return
				}

			case *AsyncReq:

				if timeout, err = gs.doCast(m.Data); err != nil {
					return
				}

			default:

				if timeout, err = gs.doInfo(m); err != nil {
					return
				}

			} // switch m.(type)

		case <-timeout:

			if timeout, err = gs.doInfo(gsTimeout); err != nil {
				return
			}

		} // select
	} // for
}

func (gs *GenServerSys) doInit(
	args ...Term) (timeout <-chan time.Time, err error) {

	err = nil
	timeout = nil

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)

			trace := make([]byte, 4096)
			n := runtime.Stack(trace, false)

			fmt.Println(time.Now().Truncate(time.Microsecond), gs.Self(),
				"crashed with reason:", r, n, "bytes stack:",
				string(trace[:n]))

			TraceCall(gs.Tracer(), gs.Self(), "Init crashed", err)
		}
		gs.initChan <- err

		if err != nil {
			timeout = nil
		}
	}()

	ts := TraceCall(gs.Tracer(), gs.Self(), traceFuncDoInit, args)

	result := gs.callbackGs.Init(args...)

	TraceCallResult(gs.Tracer(), gs.Self(), ts, traceFuncDoInit, args, result)

	switch result {

	case gsInitOk:
		return

	case gsInitTimeout:
		if gs.timeout > 0 {
			timeout = time.After(gs.timeout)
		}

	case gsStop:
		err = errors.New(gs.reason)

	default:

		switch result := result.(type) {
		case error:
			err = result
		default:
			err = fmt.Errorf("Init bad reply: %#v", result)
		}
	}

	return
}

func (gs *GenServerSys) doCall(
	req Term, replyChan From) (timeout <-chan time.Time, err error) {

	err = nil
	timeout = nil

	inCall := false

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
			if inCall {
				replyChan <- err
			}

			trace := make([]byte, 4096)
			n := runtime.Stack(trace, false)

			fmt.Println(time.Now().Truncate(time.Microsecond), gs.Self(),
				"crashed with reason:", r, n, "bytes stack:",
				string(trace[:n]))

			TraceCall(gs.Tracer(), gs.Self(), "HandleCall crashed", err)
		}

		if err != nil {
			timeout = nil
		}
	}()

	inCall = true

	ts := TraceCall(gs.Tracer(), gs.Self(), traceFuncDoCall, req)

	result := gs.callbackGs.HandleCall(req, replyChan)

	TraceCallResult(gs.Tracer(), gs.Self(), ts, traceFuncDoCall, req, result)

	inCall = false

	switch result {

	case gsCallReply:
		replyChan <- gs.reply

	case gsCallReplyOk:
		replyChan <- replyOk

	case gsCallReplyTimeout:
		replyChan <- gs.reply
		if gs.timeout > 0 {
			timeout = time.After(gs.timeout)
			gs.timeout = 0
		}

	case gsNoReply:
		return

	case gsNoReplyTimeout:
		if gs.timeout > 0 {
			timeout = time.After(gs.timeout)
			gs.timeout = 0
		}

	case gsCallStop:
		replyChan <- gs.reply
		err = errors.New(gs.reason)

	default:
		switch result := result.(type) {
		case error:
			err = result
		default:
			err = fmt.Errorf("HandleCall bad reply: %#v", result)
		}

		replyChan <- err
	}

	return
}

type asyncFunc func(req Term) Term

func (gs *GenServerSys) doCast(
	req Term) (timeout <-chan time.Time, err error) {

	return gs.doAsyncMsg(gs.callbackGs.HandleCast, traceFuncDoCast, req)
}

func (gs *GenServerSys) doInfo(
	req Term) (timeout <-chan time.Time, err error) {

	return gs.doAsyncMsg(gs.callbackGs.HandleInfo, traceFuncDoInfo, req)
}

//
// Common code for handling info or cast messages
//
func (gs *GenServerSys) doAsyncMsg(
	f asyncFunc, tag string, req Term) (timeout <-chan time.Time, err error) {

	err = nil
	timeout = nil

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)

			trace := make([]byte, 4096)
			n := runtime.Stack(trace, false)

			fmt.Println(time.Now().Truncate(time.Microsecond), gs.Self(),
				"crashed with reason:", r, n, "bytes stack:",
				string(trace[:n]))

			TraceCall(gs.Tracer(), gs.Self(), tag+" crashed", err)
		}

		if err != nil {
			timeout = nil
		}
	}()

	ts := TraceCall(gs.Tracer(), gs.Self(), tag, req)

	result := f(req)

	TraceCallResult(gs.Tracer(), gs.Self(), ts, tag, req, result)

	switch result {

	case gsNoReply:
		return

	case gsNoReplyTimeout:
		if gs.timeout > 0 {
			timeout = time.After(gs.timeout)
			gs.timeout = 0
		}

	case gsStop:
		err = errors.New(gs.reason)

	default:
		switch result := result.(type) {
		case error:
			err = result
		default:
			err = fmt.Errorf("%s bad reply: %#v", tag, result)
		}
	}

	return
}

func (gs *GenServerSys) doTerminate(reason string) {

	defer func() {
		if r := recover(); r != nil {

			trace := make([]byte, 4096)
			n := runtime.Stack(trace, false)

			fmt.Println(time.Now().Truncate(time.Microsecond), gs.Self(),
				"crashed with reason:", r, n, "bytes stack:",
				string(trace[:n]))

			TraceCall(gs.Tracer(), gs.Self(), traceFuncTerminate+" crashed", r)
		}
	}()

	ts := TraceCall(gs.Tracer(), gs.Self(), traceFuncTerminate, reason)

	gs.callbackGs.Terminate(reason)

	TraceCallResult(gs.Tracer(), gs.Self(), ts, traceFuncTerminate, reason, "")
}
