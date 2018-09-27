package stdlib

import (
	"fmt"
)

//
// Messages to registrator
//
type makePidReq struct {
	opts   *SpawnOpts
	retPid *Pid
	newPid bool
	err    error
}

type makeRefReq struct {
	replyChan chan Ref
}

type regNameReq struct {
	name Term
	pid  *Pid
}

type unregNameReq struct {
	name Term
}

type whereisReq struct {
	prefix string
	name   Term
	pid    *Pid
}

type regPrefixNameReq struct {
	prefix string
	name   Term
	pid    *Pid
}

type unregPrefixNameReq struct {
	prefix string
	name   Term
}

type whereareReq struct {
	prefix string
	regs   RegMap
}

//
// RegMap is a map of registered process names
//
type RegMap map[Term]*Pid
type nameReg struct {
	prefix string
	name   Term
}
type refReg struct {
	ref   Ref
	names []*nameReg
}

type envGs struct {
	GenServerSys

	envUID  uint32
	nextPid uint64
	ref     uint64

	// reg names
	regPrefix map[string]RegMap
	regName   RegMap

	// monitored regName, regPrefix processes
	regNameByRef map[Ref]*Pid
	regNameByPid map[*Pid]*refReg
}

//
// API
//
func mustNewEnvGs(e *Env) {
	_, err := e.GenServerStart(&envGs{envUID: e.uid})
	if err != nil {
		panic(err)
	}
}

func (e *Env) makePid(opts *SpawnOpts) (*Pid, bool, error) {

	r := &makePidReq{opts: opts}

	if _, err := e.gs.Call(r); err != nil {
		return nil, false, err
	}

	return r.retPid, r.newPid, r.err
}

func (e *Env) makeRef() Ref {

	replyChan := make(chan Ref, 1)

	if err := e.gs.Send(&makeRefReq{replyChan: replyChan}); err != nil {
		panic(err)
	}

	ref := <-replyChan

	return ref
}

func (e *Env) regPidName(prefix string, name Term, pid *Pid) (err error) {

	if name == "" {
		return NameEmptyError
	}

	if prefix == "" {
		_, err = e.gs.Call(&regNameReq{name, pid})
	} else {
		_, err = e.gs.Call(&regPrefixNameReq{prefix, name, pid})
	}
	if err != nil {
		return err
	}

	return nil
}

func (e *Env) unregPidName(prefix string, name Term) (err error) {

	if name == "" {
		return NameEmptyError
	}

	if prefix == "" {
		_, err = e.gs.Call(&unregNameReq{name})
	} else {
		_, err = e.gs.Call(&unregPrefixNameReq{prefix, name})
	}
	if err != nil {
		return err
	}

	return nil
}

func (e *Env) whereis(name Term) (*Pid, error) {
	if name == "" {
		return nil, NameEmptyError
	}

	r := &whereisReq{name: name}
	_, err := e.gs.Call(r)
	if err != nil {
		return nil, err
	}

	return r.pid, nil
}

func (e *Env) whereisPrefix(prefix string, name Term) (*Pid, error) {
	if name == "" {
		return nil, NameEmptyError
	}

	if prefix == "" {
		return nil, PrefixEmptyError
	}

	r := &whereisReq{prefix: prefix, name: name}
	_, err := e.gs.Call(r)
	if err != nil {
		return nil, err
	}

	return r.pid, nil
}

func (e *Env) whereare(prefix string) (RegMap, error) {
	if prefix == "" {
		return nil, PrefixEmptyError
	}

	r := &whereareReq{prefix: prefix}
	_, err := e.gs.Call(r)
	if err != nil {
		return nil, err
	}

	return r.regs, nil
}

// ----------------------------------------------------------------------------
// GenServer callbacks
// ----------------------------------------------------------------------------
func (gs *envGs) Init(args ...Term) Term {

	gs.SetTrapExit(true)
	// gs.SetTracer(TraceToConsole())

	return GsInitOk
}

func (gs *envGs) HandleCall(r Term, from From) Term {

	switch r := r.(type) {

	case *makePidReq:
		gs.regNewPid(r)

	case *regNameReq:
		return gs.CallReply(gs.doRegPidNameReq(r))

	case *unregNameReq:
		return gs.CallReply(gs.unregPidName(r))

	case *whereisReq:
		return gs.CallReply(gs.whereis(r))

	case *regPrefixNameReq:
		return gs.CallReply(gs.doRegPidPrefixName(r))

	case *unregPrefixNameReq:
		return gs.CallReply(gs.unregPidPrefixName(r))

	case *whereareReq:
		return gs.CallReply(gs.whereare(r))

	default:
		return fmt.Errorf("%s: unexpected call: %#v", gs.Self(), r)
	}

	return GsCallReplyOk
}

func (gs *envGs) HandleInfo(r Term) Term {

	switch r := r.(type) {

	case *makeRefReq:
		gs.regNewRef(r)

	case *MonitorDownReq:
		gs.unregNameByRef(r.MonitorRef)
	}

	return GsNoReply
}

//
// Locals
//
func (gs *envGs) regNewPid(r *makePidReq) {

	//
	// spawn + register
	//
	if r.opts.Name != nil {

		oldReg, oldPid := gs.isRegisteredPrefixName(r.opts.Prefix, r.opts.Name)
		if oldReg {
			if r.opts.returnPidIfRegistered == true {
				r.retPid = oldPid
			} else {
				r.retPid = nil
				r.err = AlreadyRegError
			}
			return
		}
	}

	pid := newPid(
		gs.nextPid+1, gs.Self().env, r.opts.UsrChanSize, r.opts.SysChanSize)

	if r.opts.Name != nil {
		if r.opts.Prefix == "" {
			gs.regPidName(r.opts.Name, pid)
		} else {
			gs.regPidPrefixName(r.opts.Prefix, r.opts.Name, pid)
		}
	}

	gs.nextPid++

	r.newPid = true
	r.retPid = pid
}

func (gs *envGs) regNewRef(r *makeRefReq) {

	r.replyChan <- gs.newRef()
}

func (gs *envGs) newRef() Ref {
	gs.ref++

	ref := Ref{
		envID: gs.envUID,
		id:    gs.ref,
	}

	return ref
}

//
// Reg name
//
func (gs *envGs) doRegPidNameReq(r *regNameReq) Term {

	if ok, _ := gs.regPidName(r.name, r.pid); ok {
		return true
	}

	return AlreadyRegError
}

func (gs *envGs) regPidName(name Term, pid *Pid) (bool, *Pid) {

	if gs.regName == nil {
		gs.regName = make(RegMap)

	}
	if gs.regNameByRef == nil {
		gs.regNameByRef = make(map[Ref]*Pid)
		gs.regNameByPid = make(map[*Pid]*refReg)
	}

	oldPid, ok := gs.regName[name]

	if !ok {
		gs.regName[name] = pid

		gs.monitorPid(pid, "", name)
		gs.dumpRegs("regPidName")

		return true, pid
	}

	return false, oldPid
}

//
// Returns BadArgError if name is not a registered name
//
func (gs *envGs) unregPidName(r *unregNameReq) Term {

	if gs.regName == nil {
		return NotRegError
	}

	pid, ok := gs.regName[r.name]
	if !ok {
		return NotRegError
	}

	delete(gs.regName, r.name)

	gs.demonitorName(pid, "", r.name)
	gs.dumpRegs("unregPidName")

	return true
}

func (gs *envGs) unregNameByRef(ref Ref) {

	if pid, ok := gs.regNameByRef[ref]; ok {

		if refReg, ok := gs.regNameByPid[pid]; ok {
			for _, nameReg := range refReg.names {
				if nameReg.prefix == "" {
					delete(gs.regName, nameReg.name)
				} else {
					delete(gs.regPrefix[nameReg.prefix], nameReg.name)
				}
			}
		}

		delete(gs.regNameByRef, ref)
		delete(gs.regNameByPid, pid)
	}

	gs.dumpRegs("unregNameByRef")
}

func (gs *envGs) whereis(r *whereisReq) Term {

	// find by name
	if r.prefix == "" {

		if gs.regName == nil {
			return NotRegError
		}

		if pid, ok := gs.regName[r.name]; ok {
			r.pid = pid
			return true
		}

		return NotRegError
	}

	// find by prefix+name
	if gs.regPrefix == nil {
		return NotRegError
	}

	if pids, ok := gs.regPrefix[r.prefix]; ok {
		if pid, pidOk := pids[r.name]; pidOk {
			r.pid = pid
			return true
		}
	}

	return NotRegError
}

func (gs *envGs) whereare(r *whereareReq) Term {
	if gs.regPrefix == nil {
		return NotRegError
	}

	if pids, ok := gs.regPrefix[r.prefix]; ok {
		r.regs = make(RegMap)
		for k, v := range pids {
			r.regs[k] = v
		}
		return true
	}

	return NotRegError
}

//
// Reg prefix + name
//
func (gs *envGs) doRegPidPrefixName(r *regPrefixNameReq) Term {

	if ok, _ := gs.regPidPrefixName(r.prefix, r.name, r.pid); ok {
		return true
	}

	return AlreadyRegError

}

func (gs *envGs) regPidPrefixName(
	prefix string, name Term, pid *Pid) (bool, *Pid) {

	gs.dumpRegs("regPidPrefixName")

	if gs.regPrefix == nil {
		gs.regPrefix = make(map[string]RegMap)

	}
	if _, ok := gs.regPrefix[prefix]; !ok {
		gs.regPrefix[prefix] = make(RegMap)
	}
	if gs.regNameByRef == nil {
		gs.regNameByRef = make(map[Ref]*Pid)
		gs.regNameByPid = make(map[*Pid]*refReg)
	}

	oldPid, ok := gs.regPrefix[prefix][name]
	if !ok {
		gs.regPrefix[prefix][name] = pid

		gs.monitorPid(pid, prefix, name)
		gs.dumpRegs("regPidPrefixName")

		return true, pid
	}

	return false, oldPid
}

func (gs *envGs) isRegisteredPrefixName(prefix string, name Term) (bool, *Pid) {
	if name == nil {
		return false, nil
	}
	if prefix == "" {
		if gs.regName == nil {
			return false, nil
		}
		pid, ok := gs.regName[name]
		return ok, pid
	}

	if gs.regPrefix == nil {
		return false, nil
	}

	if _, ok := gs.regPrefix[prefix]; !ok {
		return false, nil
	}
	pid, ok := gs.regPrefix[prefix][name]
	return ok, pid
}

//
// Returns BadArgError if name is not a registered name
//
func (gs *envGs) unregPidPrefixName(r *unregPrefixNameReq) Term {

	if gs.regPrefix == nil {
		return NotRegError
	}

	pid, ok := gs.regPrefix[r.prefix][r.name]
	if !ok {
		return NotRegError
	}

	delete(gs.regPrefix[r.prefix], r.name)

	gs.demonitorName(pid, r.prefix, r.name)
	gs.dumpRegs("unregPidPrefixName")

	return true
}

//
// Monitor pid
//
func (gs *envGs) monitorPid(pid *Pid, prefix string, name Term) {

	reg, ok := gs.regNameByPid[pid]
	if !ok {
		ref := gs.newRef()
		gs.monitorProcessPid(ref, pid)

		gs.regNameByRef[ref] = pid

		reg = &refReg{ref: ref}
	}

	reg.names = append(reg.names, &nameReg{prefix, name})

	gs.regNameByPid[pid] = reg
}

func (gs *envGs) demonitorName(pid *Pid, prefix string, name Term) {

	if refReg, ok := gs.regNameByPid[pid]; ok {

		matchedItem := -1
		for i, nameReg := range refReg.names {
			if nameReg.prefix == prefix && nameReg.name == name {
				matchedItem = i
				break
			}
		}
		if matchedItem != -1 {
			namesLen := len(refReg.names)
			if namesLen > 1 {
				refReg.names[matchedItem] = refReg.names[namesLen-1]
				refReg.names = refReg.names[:namesLen-1]
			} else {
				refReg.names = nil
			}
		}

		if len(refReg.names) == 0 {
			gs.DemonitorProcessPid(refReg.ref)
			delete(gs.regNameByRef, refReg.ref)
			delete(gs.regNameByPid, pid)
		} else {
			gs.regNameByPid[pid] = refReg
		}
	}
}

func (gs *envGs) dumpRegs(tag string) {
	if t := gs.Tracer(); t != nil {
		TraceCall(t, gs.Self(), tag, fmt.Sprintf(
			"%s:\n regName: %v\n regPrefix: %v\n"+
				" regNameByRef: %v\n regNameByPid: %v\n",
			tag, gs.regName, gs.regPrefix, gs.regNameByRef, gs.regNameByPid))
	}
}
