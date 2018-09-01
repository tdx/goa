package stdlib

//
// GenServer types
//

import (
	"time"
)

//
// GsTimeout is sent to the process if the inactivity timer has expired
//
type GsTimeout int

type gsNoReply int

//
// GsNoReplyTimeout is returned from the HandleCall/HandleCast/HandleInfo
// callbacks to indicate that an inactivity timer must be set.
// In HandleCall result to caller can be returned with Reply()
//
type GsNoReplyTimeout struct {
	Timeout time.Duration
}

// ---------------------------------------------------------------------------
// Init
// ---------------------------------------------------------------------------
//
// ok
// {ok, Timeout}
// error
//
type gsInitOk int

//
// GsInitTimeout is returned from the Init callback to indicate
// that the process initialization is successful and an inactivity
// timer must be set
//
type GsInitTimeout struct {
	Timeout time.Duration
}

// ---------------------------------------------------------------------------
// Call
// ---------------------------------------------------------------------------
//
// {reply, Reply}
// {reply, Reply, Timeout}
// noreply
// {noreply, Timeout}
// {stop, Reason, Reply}
//

type gsCallReplyOk int

//
// GsCallReply is returned from the HandleCall callback to indicate that
// the process returns result in Reply
//
type GsCallReply struct {
	Reply Term
}

//
// GsCallReplyTimeout is returned from the HandleCall callback to indicate that
// the process returns result in Reply and an inactivity timer must be set
//
type GsCallReplyTimeout struct {
	Reply   Term
	Timeout time.Duration
}

//
// GsCallStop is returned from the HandleCall callback to indicate that
// the process must be stopped
//
type GsCallStop struct {
	Reason string
	Reply  Term
}

// ---------------------------------------------------------------------------
// Cast/Info
// ---------------------------------------------------------------------------
//
// noreply
// {noreply, Timeout}
// {stop, Reason}
//

//
// GsStop is returned from the HandleCast/HandleInfo callbacks to indicate that
// the process must be stopped
//
type GsStop struct {
	Reason string
}

const (
	replyOk string = "ok"

	gsTimeout GsTimeout = 0
	// GsInitOk is returned from Init callback to indicate that initialization
	// of process is successful
	GsInitOk gsInitOk = 1
	// GsNoReply is returned from HandleCall/HandleCast/HandleInfo callbacks
	// to indicate that no result to reply to caller
	GsNoReply gsNoReply = 2
	// GsCallReplyOk is standard reply from HandleCall
	GsCallReplyOk gsCallReplyOk = 3
)
