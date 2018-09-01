package stdlib

import (
	"errors"
)

//
// ExitPidReq message sent to process when other linked process died or
//  explicitly by Exit or ExitReason
//
type ExitPidReq struct {
	From   *Pid
	Reason string
	Exit   bool
}

//
// StopPidReq is message from pid.Stop() call
//
type StopPidReq struct {
	Reason string
}

//
// Exit sends async exit request to itself process
//
func (pid *Pid) Exit(reason string) error {
	var pidFrom *Pid
	return pidFrom.ExitReason(pid, reason)
}

//
// ExitReason Sends async exit request to other process
//
func (pid *Pid) ExitReason(pidTo *Pid, reason string) error {
	if pid.Equal(pidTo) {
		return errors.New("use pid.Exit(reason) to send exit to itself")
	}
	return pid.exitReason(pidTo, reason, false)
}

func (pid *Pid) exitReason(pidTo *Pid, reason string, exit bool) error {
	return pidTo.SendSys(&ExitPidReq{pid, reason, exit})
}

//
// Stop sends sync exit request to self with reason ExitNormal
//
func (pid *Pid) Stop() (err error) {
	return pid.StopReason(ExitNormal)
}

//
// StopReason sends sync exit request to self with specified reason
//
func (pid *Pid) StopReason(reason string) (err error) {
	_, err = pid.CallSys(&StopPidReq{reason})
	return
}
