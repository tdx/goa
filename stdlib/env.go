package stdlib

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
)

//
// Env is environment for processes.
// Env used to group processes.
// Env has methods for process creation and registration.
//
type Env struct {
	uid uint32
	gs  *Pid   // GenServer for Env
	eGs *envGs // gs state to direct fast access

	syncMsgPool   sync.Pool
	sysMsgPool    sync.Pool
	replyChanPool sync.Pool
}

// ---------------------------------------------------------------------------
var env *Env // default env
var envUID uint32

func init() {
	env = NewEnv()
}

//
// NewEnv creates environment
//
func NewEnv() *Env {

	e := &Env{uid: atomic.AddUint32(&envUID, 1)}
	mustNewEnvGs(e)

	e.syncMsgPool = sync.Pool{
		New: func() interface{} {
			return &SyncReq{}
		},
	}
	e.sysMsgPool = sync.Pool{
		New: func() interface{} {
			return &SysReq{}
		},
	}
	e.replyChanPool = sync.Pool{
		New: func() interface{} {
			return make(chan Term, 1)
		},
	}

	return e
}

//
// Stat dump env statistics
//
func Stat(w io.Writer) {
	if env.eGs != nil {
		env.eGs.Stat(w)
	} else {
		fmt.Fprintln(w, "no gen server yet")
	}
}

func (e *Env) id() uint32 {
	return e.uid
}

//
// messages pools
//
func (e *Env) getSyncMsg() *SyncReq {
	return e.syncMsgPool.Get().(*SyncReq)
}

func (e *Env) putSyncMsg(m *SyncReq) {
	e.syncMsgPool.Put(m)
}

func (e *Env) getSysMsg() *SysReq {
	return e.sysMsgPool.Get().(*SysReq)
}

func (e *Env) putSysMsg(m *SysReq) {
	e.sysMsgPool.Put(m)
}

func (e *Env) getReplyChan() chan Term {
	return e.replyChanPool.Get().(chan Term)
}

func (e *Env) putReplyChan(c chan Term) {
	e.replyChanPool.Put(c)
}

//
func (e *Env) spawnObjOpts(
	gp GenProc, opts *SpawnOpts, args ...Term) (*Pid, error) {

	if opts.UsrChanSize == 0 {
		opts.UsrChanSize = 16
	}
	if opts.SysChanSize == 0 {
		opts.SysChanSize = 8
	}

	if opts.link && opts.linkPid == nil {
		return nil, NilPidError
	}

	if opts.returnPidIfRegistered && opts.Name == nil {
		return nil, NameEmptyError
	}

	pid, newPid, err := e.newPid(opts)
	if err != nil {
		return nil, err
	}

	if opts.returnPidIfRegistered && !newPid {
		return pid, err
	}

	gp.setPid(pid)
	gp.InitPrepare()
	gp.SetTracer(opts.tracer)

	go gp.Run(gp, opts, args...)

	if err := gp.InitAck(); err != nil {
		return nil, err
	}

	return pid, nil
}

func (e *Env) newPid(opts *SpawnOpts) (*Pid, bool, error) {

	if e.gs != nil {
		return e.makePid(opts)
	}

	e.gs = e.makeEnvPid(opts)

	return e.gs, true, nil
}

func (e *Env) makeEnvPid(opts *SpawnOpts) *Pid {
	return e.makeEnvPidOpts(opts)
}

func (e *Env) makeEnvPidOpts(opts *SpawnOpts) *Pid {
	return newPid(0, e, opts.UsrChanSize, opts.SysChanSize)
}

//
//
//
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
