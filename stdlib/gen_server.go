package stdlib

import (
	"errors"
)

//
// From is a channel to reply to the caller of the process
//
type From chan<- Term

//
// GenServer is an interface for callbacks functions of the process
//
type GenServer interface {
	GenProc

	Init(args ...Term) Term
	HandleCall(req Term, from From) (result Term)
	HandleCast(req Term) (result Term)
	HandleInfo(req Term) (result Term)
	Terminate(reason string)

	// private
	setCallback(gs GenServer)
}

//
// GenServerStart start GenServer process in default environment
//
func GenServerStart(
	gs GenServer, args ...Term) (*Pid, error) {

	return env.GenServerStartOpts(gs, NewSpawnOpts(), args...)
}

//
// GenServerStart start GenServer process in specified environment
//
func (e *Env) GenServerStart(
	gs GenServer, args ...Term) (*Pid, error) {

	return e.GenServerStartOpts(gs, NewSpawnOpts(), args...)
}

//
// GenServerStartLink start GenServer process in default environment
//  and links it to called process
//
func (pid *Pid) GenServerStartLink(
	gs GenServer, opts *SpawnOpts, args ...Term) (*Pid, error) {

	if pid == nil {
		return nil, NilPidError
	}
	if opts == nil {
		opts = NewSpawnOpts()
	}

	opts = opts.WithLinkTo(pid)

	return pid.env.GenServerStartOpts(gs, opts, args...)
}

//
// GenServerStartOpts start GenServer process in default environment
//  with given options
//
func GenServerStartOpts(
	gs GenServer, opts *SpawnOpts, args ...Term) (*Pid, error) {

	return env.GenServerStartOpts(gs, opts, args...)
}

//
// GenServerStartOpts start GenServer process in specified environment
//  with given options
//
func (e *Env) GenServerStartOpts(
	gs GenServer, opts *SpawnOpts, args ...Term) (*Pid, error) {

	if gs == nil {
		return nil, errors.New("GenServer parameter is nil")
	}

	if opts == nil {
		opts = NewSpawnOpts()
	}

	gs.setCallback(gs)

	return e.SpawnObj(gs, opts, args...)
}
