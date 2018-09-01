package stdlib

//
// MonitorPidReq is a message to monitor process
//
type MonitorPidReq struct {
	MonitorRef Ref
	PidFrom    *Pid
}

//
// DemonitorPidReq is a message to demonitor process
//
type DemonitorPidReq struct {
	MonitorRef Ref
}

//
// MonitorDownReq is a message sent when monitored process died
//
type MonitorDownReq struct {
	MonitorRef Ref
	PidFrom    *Pid
	Reason     string
}

func (pidFrom *Pid) monitorDown(pidTo *Pid, ref Ref, reason string) {
	_ = pidTo.SendSys(&MonitorDownReq{ref, pidFrom, reason})
}
