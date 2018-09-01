package stdlib

import (
	// "fmt"
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
	gs  *Pid // GenServer for Env

	syncMsgPool   sync.Pool
	sysMsgPool    sync.Pool
	replyChanPool sync.Pool

	initTimeoutPool sync.Pool

	callReplyPool        sync.Pool
	callReplyTimeoutPool sync.Pool
	callStopPool         sync.Pool

	noReplyTimeoutPool sync.Pool
	stopPool           sync.Pool
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

	e.initTimeoutPool = sync.Pool{
		New: func() interface{} {
			return &GsInitTimeout{}
		},
	}
	e.callReplyPool = sync.Pool{
		New: func() interface{} {
			return &GsCallReply{}
		},
	}
	e.callReplyTimeoutPool = sync.Pool{
		New: func() interface{} {
			return &GsCallReplyTimeout{}
		},
	}
	e.callStopPool = sync.Pool{
		New: func() interface{} {
			return &GsCallStop{}
		},
	}
	e.noReplyTimeoutPool = sync.Pool{
		New: func() interface{} {
			return &GsNoReplyTimeout{}
		},
	}
	e.stopPool = sync.Pool{
		New: func() interface{} {
			return &GsStop{}
		},
	}

	return e
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

func (e *Env) getInitTimeout() *GsInitTimeout {
	return e.initTimeoutPool.Get().(*GsInitTimeout)
}

func (e *Env) putInitTimeout(r *GsInitTimeout) {
	e.initTimeoutPool.Put(r)
}

func (e *Env) getCallReply() *GsCallReply {
	return e.callReplyPool.Get().(*GsCallReply)
}

func (e *Env) putCallReply(r *GsCallReply) {
	e.callReplyPool.Put(r)
}

func (e *Env) getCallReplyTimeout() *GsCallReplyTimeout {
	return e.callReplyTimeoutPool.Get().(*GsCallReplyTimeout)
}

func (e *Env) putCallReplyTimeout(r *GsCallReplyTimeout) {
	e.callReplyTimeoutPool.Put(r)
}

func (e *Env) getCallStop() *GsCallStop {
	return e.callStopPool.Get().(*GsCallStop)
}

func (e *Env) putCallStop(r *GsCallStop) {
	e.callStopPool.Put(r)
}

func (e *Env) getNoReplyTimeout() *GsNoReplyTimeout {
	return e.noReplyTimeoutPool.Get().(*GsNoReplyTimeout)
}

func (e *Env) putNoReplyTimeout(r *GsNoReplyTimeout) {
	e.noReplyTimeoutPool.Put(r)
}

func (e *Env) getStop() *GsStop {
	return e.stopPool.Get().(*GsStop)
}

func (e *Env) putStop(r *GsStop) {
	e.stopPool.Put(r)
}

//
// Returns: pid, isNewPid, error
//
func (e *Env) spawnObjOpts(
	gp GenProc, opts *SpawnOpts, args ...Term) (*Pid, error) {

	if opts.UsrChanSize == 0 {
		opts.UsrChanSize = 256
	}
	if opts.SysChanSize == 0 {
		opts.SysChanSize = 128
	}

	if opts.link && opts.linkPid == nil {
		return nil, NilPidError
	}

	if opts.returnPidIfRegistered && opts.Name == nil {
		return nil, NameEmptyError
	}

	pid, newPid, err := e.newPid(opts)
	// fmt.Println("newPid:", pid)
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

	e.gs = e.makeEnvPid()

	return e.gs, true, nil
}

func (e *Env) makeEnvPid() *Pid {
	return e.makeEnvPidOpts(
		NewSpawnOpts().
			WithUsrChannelSize(1024).
			WithSysChannelSize(512))
}

func (e *Env) makeEnvPidOpts(opts *SpawnOpts) *Pid {

	return newPid(0, e, opts.UsrChanSize, opts.SysChanSize)
}
