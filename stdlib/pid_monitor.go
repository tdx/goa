package stdlib

//
// Monitors
//
func (pid *Pid) monitorMe(mPid *Pid, ref Ref) {
	pid.mu.Lock()
	defer pid.mu.Unlock()

	if pid.monitors == nil {
		pid.monitors = make(map[Ref]*Pid)
	}
	pid.monitors[ref] = mPid
}

func (pid *Pid) demonitorMe(ref Ref) {
	pid.mu.Lock()
	defer pid.mu.Unlock()

	if pid.monitors == nil {
		return
	}
	delete(pid.monitors, ref)
}

func (pid *Pid) monitorByMe(mPid *Pid, ref Ref) {
	pid.mu.Lock()
	defer pid.mu.Unlock()

	if pid.monitorsByMe == nil {
		pid.monitorsByMe = make(map[Ref]*Pid)
	}
	pid.monitorsByMe[ref] = mPid

}

func (pid *Pid) demonitorByMe(ref Ref) *Pid {
	pid.mu.Lock()
	defer pid.mu.Unlock()

	if pid.monitorsByMe == nil {
		return nil
	}

	if mPid, ok := pid.monitorsByMe[ref]; ok {
		delete(pid.monitorsByMe, ref)
		return mPid
	}

	return nil
}

//
func (pid *Pid) onStop(reason string) {

	pid.mu.Lock()

	monitors := pid.monitors
	pid.monitors = nil
	pid.monitorsByMe = nil

	pid.mu.Unlock()

	for ref, mPid := range monitors {
		pid.monitorDown(mPid, ref, reason)
	}
}

//
// MonitorDownReq is a message sent when monitored process died
//
type MonitorDownReq struct {
	MonitorRef Ref
	PidFrom    *Pid
	Reason     string
}

func (pid *Pid) monitorDown(pidTo *Pid, ref Ref, reason string) {
	_ = pidTo.Send(&MonitorDownReq{ref, pid, reason})
}
