package stdlib

//
// GenProcFunc is a type for loop function of the process
//
type GenProcFunc func(gp GenProc, args ...Term) error

//
// GenProc interface
//
type GenProc interface {
	//
	// Handle sys messages for process
	//
	HandleSysMsg(msg Term) error

	//
	// Process flag
	//
	SetTrapExit(flag bool)
	TrapExit() bool

	//
	// Returns pid of the process
	//
	Self() *Pid
	setPid(pid *Pid)

	//
	// Runs f, transforms return value into ExitPidReq message
	//
	Run(gp GenProc, opts *SpawnOpts, args ...Term)
	GenProcLoop(args ...Term) error
	//
	// Waits ack from f
	//
	InitPrepare()
	InitAck() error
	//
	// Links
	//
	Link(*Pid)
	Unlink(*Pid)
	//
	// Monitors
	//
	MonitorProcessPid(*Pid) Ref
	DemonitorProcessPid(Ref)
	//
	// Tracer
	//
	SetTracer(Tracer)
	Tracer() Tracer
}
