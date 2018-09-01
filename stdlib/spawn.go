package stdlib

//
// Spawn makes new GenProc object and runs it's GenProcLoop
//
func Spawn(
	f GenProcFunc, args ...Term) (*Pid, error) {

	return env.Spawn(f, args...)
}

//
// Spawn makes new GenProc object and runs it's GenProcLoop
//
func (e *Env) Spawn(
	f GenProcFunc, args ...Term) (*Pid, error) {

	return e.SpawnWithOpts(f, NewSpawnOpts(), args...)
}

//
// SpawnWithOpts makes new GenProc object and runs it's GenProcLoop
//
func SpawnWithOpts(
	f GenProcFunc, opts *SpawnOpts, args ...Term) (*Pid, error) {

	return env.SpawnWithOpts(f, opts, args...)
}

//
// SpawnWithOpts makes new GenProc object and runs it's GenProcLoop
//
func (e *Env) SpawnWithOpts(
	f GenProcFunc, opts *SpawnOpts, args ...Term) (*Pid, error) {

	if opts == nil {
		opts = NewSpawnOpts()
	}

	return e.spawnObjOpts(NewGenProcSys(f), opts, args...)
}

//
// SpawnLink makes new GenProc object, runs it's GenProcLoop and link processes
//
func (pidFrom *Pid) SpawnLink(
	f GenProcFunc, args ...Term) (*Pid, error) {

	return pidFrom.SpawnOptsLink(f, NewSpawnOpts(), args...)
}

//
// SpawnOptsLink makes new GenProc object,
//  runs it's GenProcLoop and link processes
//
func (pidFrom *Pid) SpawnOptsLink(
	f GenProcFunc, opts *SpawnOpts, args ...Term) (*Pid, error) {

	if pidFrom == nil {
		return nil, NilPidError
	}

	if opts == nil {
		opts = NewSpawnOpts()
	}

	opts = opts.WithLinkTo(pidFrom)

	return pidFrom.env.spawnObjOpts(NewGenProcSys(f), opts, args...)
}

//
// SpawnObj - function to exdend default GenProc object
//
func SpawnObj(
	gp GenProc, opts *SpawnOpts, args ...Term) (*Pid, error) {

	return env.SpawnObj(gp, opts, args...)
}

//
// SpawnObj - function to exdend default GenProc object
//
func (e *Env) SpawnObj(
	gp GenProc, opts *SpawnOpts, args ...Term) (*Pid, error) {

	if opts == nil {
		opts = NewSpawnOpts()
	}

	return e.spawnObjOpts(gp, opts, args...)
}
