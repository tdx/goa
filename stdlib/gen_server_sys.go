package stdlib

import (
	"errors"
	"fmt"
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
	return GsInitOk
}

//
// HandleCall handles incoming messages from `pid.Call(data)`
//
func (gs *GenServerSys) HandleCall(req Term, from From) Term {
	return GsCallReplyOk
}

//
// HandleCast handles incoming messages from `pid.Cast(data)`
//
func (gs *GenServerSys) HandleCast(req Term) Term {
	return GsNoReply
}

//
// HandleInfo handles timeouts and system messages
//
func (gs *GenServerSys) HandleInfo(req Term) Term {
	return GsNoReply
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
// InitTimeout makes InitTimeout reply
//
func (gs *GenServerSys) InitTimeout(timeout time.Duration) *GsInitTimeout {

	r := gs.Self().env.getInitTimeout()
	r.Timeout = timeout
	return r
}

//
// CallReply makes CallReply reply
//
func (gs *GenServerSys) CallReply(reply Term) *GsCallReply {
	r := gs.Self().env.getCallReply()
	r.Reply = reply
	return r
}

//
// CallReplyTimeout makes CallReplyTimeout reply
//
func (gs *GenServerSys) CallReplyTimeout(
	reply Term, timeout time.Duration) *GsCallReplyTimeout {

	r := gs.Self().env.getCallReplyTimeout()
	r.Reply = reply
	r.Timeout = timeout
	return r
}

//
// CallStop makes CallStop reply
//
func (gs *GenServerSys) CallStop(reason string, reply Term) *GsCallStop {
	r := gs.Self().env.getCallStop()
	r.Reason = reason
	r.Reply = reply
	return r
}

//
// Stop makes Stop reply
//
func (gs *GenServerSys) Stop(reason string) *GsStop {
	r := gs.Self().env.getStop()
	r.Reason = reason
	return r
}

//
// NoReplyTimeout makes NoReplyTimeout reply
//
func (gs *GenServerSys) NoReplyTimeout(
	timeout time.Duration) *GsNoReplyTimeout {

	r := gs.Self().env.getNoReplyTimeout()
	r.Timeout = timeout
	return r
}

//
// GenProcLoop is a main gen_server loop function
//
func (gs *GenServerSys) GenProcLoop(args ...Term) (err error) {

	pid := gs.Self()

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
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

			case *SysReq:

				if timeout, err = gs.doInfo(m.Data); err != nil {
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
			gs.initChan <- err
			TraceCall(gs.Tracer(), gs.Self(), "Init crashed", err)
		} else {
			gs.initChan <- err
		}

		if err != nil {
			timeout = nil
		}
	}()

	TraceCall(gs.Tracer(), gs.Self(), traceFuncDoInit, args)
	ts := time.Now()
	result := gs.callbackGs.Init(args...)
	TraceCallResult(gs.Tracer(), gs.Self(), &ts, traceFuncDoInit, args, result)

	switch result := result.(type) {

	case gsInitOk:
		return

	case *GsInitTimeout:
		if result.Timeout > 0 {
			timeout = time.After(result.Timeout)
		}
		gs.Self().env.putInitTimeout(result)

	case *GsStop:
		err = errors.New(result.Reason)
		gs.Self().env.putStop(result)

	case error:
		err = result

	default:
		err = fmt.Errorf("Init bad reply: %#v", result)
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

			TraceCall(
				gs.Tracer(), gs.Self(), "HandleCall crashed", err)
		}

		if err != nil {
			timeout = nil
		}
	}()

	inCall = true

	TraceCall(gs.Tracer(), gs.Self(), traceFuncDoCall, req)
	ts := time.Now()

	result := gs.callbackGs.HandleCall(req, replyChan)

	TraceCallResult(gs.Tracer(), gs.Self(), &ts, traceFuncDoCall, req, result)

	inCall = false

	switch result := result.(type) {

	case *GsCallReply:
		replyChan <- result.Reply
		gs.Self().env.putCallReply(result)

	case gsCallReplyOk:
		replyChan <- replyOk

	case *GsCallReplyTimeout:
		replyChan <- result.Reply
		if result.Timeout > 0 {
			timeout = time.After(result.Timeout)
		}
		gs.Self().env.putCallReplyTimeout(result)

	case gsNoReply:
		return

	case *GsNoReplyTimeout:
		if result.Timeout > 0 {
			timeout = time.After(result.Timeout)
		}
		gs.Self().env.putNoReplyTimeout(result)

	case *GsCallStop:
		replyChan <- result.Reply
		err = errors.New(result.Reason)
		gs.Self().env.putCallStop(result)

	case error:
		replyChan <- result
		err = result

	default:
		err = fmt.Errorf("HandleCall bad reply: %#v", result)
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
			TraceCall(gs.Tracer(), gs.Self(), tag+" crashed", err)
		}

		if err != nil {
			timeout = nil
		}
	}()

	TraceCall(gs.Tracer(), gs.Self(), tag, req)
	ts := time.Now()

	result := f(req)

	TraceCallResult(gs.Tracer(), gs.Self(), &ts, tag, req, result)

	switch result := result.(type) {

	case gsNoReply:
		return

	case *GsNoReplyTimeout:
		if result.Timeout > 0 {
			timeout = time.After(result.Timeout)
		}
		gs.Self().env.putNoReplyTimeout(result)

	case *GsStop:
		err = errors.New(result.Reason)
		gs.Self().env.putStop(result)

	case error:
		err = result

	default:
		err = fmt.Errorf("%s bad reply: %#v", tag, result)
	}

	return
}

func (gs *GenServerSys) doTerminate(reason string) {

	defer func() {
		if r := recover(); r != nil {
			TraceCall(gs.Tracer(), gs.Self(), traceFuncTerminate+" crashed", r)
		}
	}()

	TraceCall(gs.Tracer(), gs.Self(), traceFuncTerminate, reason)
	ts := time.Now()

	gs.callbackGs.Terminate(reason)

	TraceCallResult(gs.Tracer(), gs.Self(), &ts, traceFuncTerminate, reason, "")
}
