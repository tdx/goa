package stdlib

//
// Monitors
//

// MonitorDownFunc is a callback for handle MonitorDown
type MonitorDownFunc func(ref Ref, reason string)

// RegisterMonitorDownFunc registers callback function
func (pid *Pid) RegisterMonitorDownFunc(fn MonitorDownFunc) {
	pid.monitorDownFunc = fn
}

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

// func (pid *Pid) demonitorByMe(ref Ref) *Pid {
func (pid *Pid) demonitorByMe(onStop bool, ref Ref, reason string) *Pid {

	var (
		ok   bool
		mPid *Pid
	)

	pid.mu.RLock()
	monByMe := pid.monitorsByMe
	if monByMe != nil {
		mPid, ok = pid.monitorsByMe[ref]
	}
	pid.mu.RUnlock()

	// fmt.Println(pid, "demontorByMe:", ok, mPid, pid.monitorDownFunc)

	if monByMe == nil {
		if onStop && pid.monitorDownFunc != nil {
			pid.monitorDownFunc(ref, reason)
		}
		return nil
	}

	if ok {
		pid.mu.Lock()
		delete(pid.monitorsByMe, ref)
		pid.mu.Unlock()
	}

	if onStop && pid.monitorDownFunc != nil {
		pid.monitorDownFunc(ref, reason)
	}

	return mPid
}

//
func (pid *Pid) onStop(reason string) {

	pid.mu.Lock()

	monitors := pid.monitors
	pid.monitors = nil
	pid.monitorsByMe = nil

	pid.mu.Unlock()

	// fmt.Println(pid, "onStop")

	for ref, mPid := range monitors {
		pid.monitorDown(mPid, ref, reason)
	}
}

func (pid *Pid) monitorDown(pidTo *Pid, ref Ref, reason string) {
	// fmt.Println(pidTo, "monitorDown:", reason, ref)
	pidTo.demonitorByMe(true, ref, reason)
}
