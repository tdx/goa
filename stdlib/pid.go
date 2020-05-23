package stdlib

import (
	"fmt"
	"sync"
)

//
// Pid encapsulates process identificator and
// channels to communicate with process
//
type Pid struct {
	id  uint64
	env *Env

	usrChan  chan Term
	sysChan  chan Term
	exitChan chan bool

	monitorDownFunc MonitorDownFunc

	mu           sync.RWMutex
	monitorsByMe map[Ref]*Pid
	monitors     map[Ref]*Pid
}

func newPid(id uint64, e *Env, usrChanSize, sysChanSize int) *Pid {

	pid := &Pid{
		id:       id,
		env:      e,
		usrChan:  make(chan Term, usrChanSize),
		sysChan:  make(chan Term, sysChanSize),
		exitChan: make(chan bool),
	}

	return pid
}

//
// ID returns process identificator
//
func (pid *Pid) ID() uint64 {
	if pid == nil {
		return 0
	}

	return pid.id
}

//
// String returns string presentation of pid
//
func (pid *Pid) String() string {
	if pid == nil {
		return "<nil>"
	}

	return fmt.Sprintf("<0.%d.%d>", pid.env.id(), pid.id)
}

//
// Equal compares two pids
//
func (pid *Pid) Equal(pid2 *Pid) bool {
	if pid == nil && pid2 == nil {
		return true
	}
	if pid == nil || pid2 == nil {
		return false
	}

	return pid.id == pid2.id && pid.env.id() == pid2.env.id()
}

//
// Alive verifies that the process is alive
//
func (pid *Pid) Alive() error {
	// fmt.Println(pid, "alive")
	if pid == nil {
		return NilPidError
	}

	// fmt.Println(pid, "alive: exitChan=", pid.exitChan)

	select {
	case <-pid.exitChan:
		// fmt.Println(pid, "alive: case NoProcError")
		return NoProcError
	default:
		// fmt.Println(pid, "alive: case default")
		return nil
	}
}

//
// GetUsrChannel returns usr channel
//
func (pid *Pid) GetUsrChannel() <-chan Term {
	if pid == nil {
		return nil
	}

	return pid.usrChan
}

//
// GetSysChannel returns sys channel
//
func (pid *Pid) GetSysChannel() <-chan Term {
	if pid == nil {
		return nil
	}

	return pid.sysChan
}
