package stdlib

import (
	"fmt"
	"sync"
	"sync/atomic"
)

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

	// mu guard all envGs fields
	mu sync.RWMutex

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
	eGs := &envGs{envUID: e.uid}
	_, err := e.GenServerStart(eGs)
	if err != nil {
		panic(err)
	}
	e.eGs = eGs
}

//
func (e *Env) makePid(opts *SpawnOpts) (*Pid, bool, error) {

	if opts == nil {
		opts = NewSpawnOpts()
	}

	e.eGs.mu.Lock()
	defer e.eGs.mu.Unlock()

	return e.eGs.regNewPid(opts)
}

//
func (e *Env) makeRef() Ref {

	return e.eGs.newRef()
}

//
func (e *Env) regPidName(prefix string, name Term, pid *Pid) error {

	if name == "" {
		return NameEmptyError
	}

	e.eGs.mu.Lock()
	defer e.eGs.mu.Unlock()

	if prefix == "" {
		if isNew, _ := e.eGs.regPidName(name, pid); isNew {
			return nil
		}
		return AlreadyRegError
	}

	if isNew, _ := e.eGs.regPidPrefixName(prefix, name, pid); isNew {
		return nil
	}

	return AlreadyRegError
}

//
func (e *Env) unregPidName(prefix string, name Term) error {

	if name == "" {
		return NameEmptyError
	}

	e.eGs.mu.Lock()
	defer e.eGs.mu.Unlock()

	if prefix == "" {
		return e.eGs.unregPidName(name)
	}

	return e.eGs.unregPidPrefixName(prefix, name)
}

func (e *Env) whereis(name Term) (pid *Pid, err error) {
	if name == "" {
		return nil, NameEmptyError
	}

	e.eGs.mu.RLock()
	pid, err = e.eGs.whereis("", name)
	e.eGs.mu.RUnlock()

	return
}

//
func (e *Env) whereisPrefix(prefix string, name Term) (pid *Pid, err error) {
	if name == "" {
		return nil, NameEmptyError
	}

	if prefix == "" {
		return nil, PrefixEmptyError
	}

	e.eGs.mu.RLock()
	pid, err = e.eGs.whereis(prefix, name)
	e.eGs.mu.RUnlock()

	return
}

//
func (e *Env) whereare(prefix string) (RegMap, error) {
	if prefix == "" {
		return nil, PrefixEmptyError
	}

	e.eGs.mu.RLock()
	regs, err := e.eGs.whereare(prefix)
	e.eGs.mu.RUnlock()

	return regs, err
}

// ----------------------------------------------------------------------------
// GenServer callbacks
// ----------------------------------------------------------------------------
func (gs *envGs) Init(args ...Term) Term {

	gs.SetTrapExit(true)
	// gs.SetTracer(TraceToConsole())

	return gs.InitOk()
}

//
func (gs *envGs) HandleInfo(r Term) Term {

	switch r := r.(type) {

	case *MonitorDownReq:
		gs.unregNameByRef(r.MonitorRef)
	}

	return gs.NoReply()
}

//
func (gs *envGs) Terminate(reason string) {
	fmt.Println("env_gs", gs.Self().String(), "terminated:", reason)
}

//
// Locals
//
func (gs *envGs) regNewPid(opts *SpawnOpts) (pid *Pid, isNewPid bool, err error) {
	//
	// spawn + register
	//
	if opts.Name != nil {
		oldReg, oldPid := gs.isRegisteredPrefixName(opts.Prefix, opts.Name)
		if oldReg {
			if opts.returnPidIfRegistered == true {
				pid = oldPid
			} else {
				pid = nil
				err = AlreadyRegError
			}
			return
		}
	}

	pid = newPid(
		gs.nextPid+1, gs.Self().env, opts.UsrChanSize, opts.SysChanSize)

	if opts.Name != nil {
		if opts.Prefix == "" {
			gs.regPidName(opts.Name, pid)
		} else {
			gs.regPidPrefixName(opts.Prefix, opts.Name, pid)
		}
	}

	gs.nextPid++

	isNewPid = true
	err = nil

	return
}

//
func (gs *envGs) newRef() Ref {
	return Ref{
		envID: gs.envUID,
		id:    atomic.AddUint64(&gs.ref, 1),
	}
}

//
// Reg name
//
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
func (gs *envGs) unregPidName(name Term) error {

	if gs.regName == nil {
		return NotRegError
	}

	pid, ok := gs.regName[name]
	if !ok {
		return NotRegError
	}

	delete(gs.regName, name)

	gs.demonitorName(pid, "", name)
	gs.dumpRegs("unregPidName")

	return nil
}

func (gs *envGs) unregNameByRef(ref Ref) {

	gs.mu.Lock()
	defer gs.mu.Unlock()

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

func (gs *envGs) whereis(prefix string, name Term) (*Pid, error) {

	// find by name
	if prefix == "" {

		if gs.regName == nil {
			return nil, NotRegError
		}

		if pid, ok := gs.regName[name]; ok {
			return pid, nil
		}

		return nil, NotRegError
	}

	// find by prefix+name
	if gs.regPrefix == nil {
		return nil, NotRegError
	}

	if pids, ok := gs.regPrefix[prefix]; ok {
		if pid, pidOk := pids[name]; pidOk {
			return pid, nil
		}
	}

	return nil, NotRegError
}

//
func (gs *envGs) whereare(prefix string) (RegMap, error) {
	if gs.regPrefix == nil {
		return nil, NotRegError
	}

	if pids, ok := gs.regPrefix[prefix]; ok {
		regs := make(RegMap)
		for k, v := range pids {
			regs[k] = v
		}
		return regs, nil
	}

	return nil, NotRegError
}

//
// Reg prefix + name
//
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
func (gs *envGs) unregPidPrefixName(prefix string, name Term) error {

	if gs.regPrefix == nil {
		return NotRegError
	}

	pid, ok := gs.regPrefix[prefix][name]
	if !ok {
		return NotRegError
	}

	delete(gs.regPrefix[prefix], name)

	gs.demonitorName(pid, prefix, name)
	gs.dumpRegs("unregPidPrefixName")

	return nil
}

//
// Monitor pid
//
func (gs *envGs) monitorPid(pid *Pid, prefix string, name Term) {

	reg, ok := gs.regNameByPid[pid]
	if !ok {
		ref := gs.newRef()
		pid.monitorMe(gs.pid, ref)
		gs.pid.monitorByMe(pid, ref)

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
