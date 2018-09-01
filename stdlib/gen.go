package stdlib

import (
// "fmt"
)

//
// Term is a type for any values
//
type Term interface{}

//
// AsyncReq is an async message to process
//
type AsyncReq struct {
	Data Term
}

//
// SyncReq is a sync message to process
//
type SyncReq struct {
	Data      Term
	ReplyChan chan<- Term
}

//
// SysReq is a system message to process
//
type SysReq struct {
	Data      Term
	ReplyChan chan<- Term
}

type callType int

const (
	callTypeSys    = 0
	callTypeUsr    = 1
	callTypeSysUsr = 2
)

//
// Send sends async message to the usr channel of the process
//
func (pid *Pid) Send(data Term) error {
	return pid.send(callTypeUsr, data)
}

//
// SendSys sends sys async message to the sys channel of the process
//
func (pid *Pid) SendSys(data Term) error {
	return pid.send(callTypeSys, data)
}

//
// SendInfo sends sys async message to the usr channel of the process
//
func (pid *Pid) SendInfo(data Term) error {
	return pid.send(callTypeSysUsr, data)
}

func (pid *Pid) send(ct callType, data Term) (err error) {

	defer func() {
		if r := recover(); r != nil {
			err = NoProcError
		}
	}()

	if err = pid.Alive(); err != nil {
		return
	}

	switch ct {

	case callTypeUsr:
		err = pid.sendUsr(&AsyncReq{data})

	case callTypeSys:
		err = pid.sendSys(&SysReq{data, nil})

	case callTypeSysUsr:
		err = pid.sendUsr(&SysReq{data, nil})

	}

	return
}

//
// Call sends sync message to the usr channel of ther process
//
func (pid *Pid) Call(data Term) (Term, error) {
	return pid.call(callTypeUsr, data)
}

//
// CallSys sends sync sys message to the sys channel of ther process
//
func (pid *Pid) CallSys(data Term) (Term, error) {
	return pid.call(callTypeSys, data)
}

func (pid *Pid) call(ct callType, data Term) (reply Term, err error) {

	defer func() {
		if r := recover(); r != nil {
			reply = nil
			err = NoProcError
		}
	}()

	if err = pid.Alive(); err != nil {
		return
	}

	replyChan := pid.env.getReplyChan()
	defer pid.env.putReplyChan(replyChan)

	switch ct {

	case callTypeSys:
		r := pid.env.getSysMsg()
		defer pid.env.putSysMsg(r)

		r.Data = data
		r.ReplyChan = replyChan
		err = pid.sendSys(r)

	case callTypeUsr:
		r := pid.env.getSyncMsg()
		defer pid.env.putSyncMsg(r)

		r.Data = data
		r.ReplyChan = replyChan
		err = pid.sendUsr(r)
	}

	if err != nil {
		return nil, err
	}

	// fmt.Printf("%s: before wait reply: %#v\n", pid, data)
	processExit := false

	select {
	case reply = <-replyChan:
		if reply == nil {
			processExit = true
		}
	case <-pid.exitChan:
		processExit = true
	}

	// fmt.Printf("%s: after wait reply: %#v - %#v\n", pid, data, reply)

	if processExit == true {

		close(pid.sysChan)
		close(pid.usrChan)

		switch data.(type) {
		case *StopPidReq:
			return true, nil
		}
		return nil, NoProcError
	}

	switch err := reply.(type) {
	case error:
		return nil, err
	}

	return reply, nil
}

func (pid *Pid) sendUsr(r Term) (err error) {
	select {
	case pid.usrChan <- r:
		return nil
	default:
		return ChannelFullError
	}
}

func (pid *Pid) sendSys(r *SysReq) (err error) {
	select {
	case pid.sysChan <- r:
		return nil
	default:
		return ChannelFullError
	}
}
