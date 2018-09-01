package stdlib

//
// LinkPidReq is a message to link pid
//
type LinkPidReq struct {
	Pid *Pid
}

//
// UnlinkPidReq is a message to unlink pid
//
type UnlinkPidReq struct {
	Pid *Pid
}

func (pid *Pid) link(pidTo *Pid) error {
	if pid.Equal(pidTo) {
		return nil
	}

	return pidTo.SendSys(&LinkPidReq{pid})
}

func (pid *Pid) unlink(pidTo *Pid) error {
	if pid.Equal(pidTo) {
		return nil
	}
	return pidTo.SendSys(&UnlinkPidReq{pid})
}

//
// ProcessLinksReq is a message Links of the process
//
type ProcessLinksReq struct {
	Links []*Pid
}

//
// ProcessLinks returns process links
//
func (pid *Pid) ProcessLinks() ([]*Pid, error) {
	r := &ProcessLinksReq{}
	_, err := pid.CallSys(r)
	return r.Links, err
}
