package stdlib

//
// GenServer callback return values
//

//
// GsReply is a type for values returned from GenServer callbacks
//
type GsReply int

const (
	replyOk string = "ok"

	gsInitOk GsReply = iota
	gsInitTimeout
	gsNoReply
	gsNoReplyTimeout
	gsCallReply
	gsCallReplyOk
	gsCallReplyTimeout
	gsCallStop
	gsStop
)

// GsTimeout is a timeout message type
type GsTimeout int

const (
	gsTimeout GsTimeout = 0
)
