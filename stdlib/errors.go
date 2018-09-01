package stdlib

type noProcError int
type nilPidError int
type chanFullError int
type notRegError int
type alreadyRegError int
type nameEmptyError int
type prefixEmptyError int

// Errors constants
const (
	NoProcError      noProcError      = 1
	NilPidError      nilPidError      = 2
	ChannelFullError chanFullError    = 3
	AlreadyRegError  alreadyRegError  = 5
	NotRegError      notRegError      = 6
	NameEmptyError   nameEmptyError   = 7
	PrefixEmptyError prefixEmptyError = 8

	NoProc string = "no_proc"
)

//
// IsNoProcError checks if error is of type of noProcError
//
func IsNoProcError(err error) bool {
	_, ok := err.(noProcError)
	return ok
}

//
// Returns string representaion of noProcError
//
func (e noProcError) Error() string {
	return NoProc
}

//
// IsNilPidError checks if error is a NilPidError
//
func IsNilPidError(err error) bool {
	_, ok := err.(nilPidError)
	return ok
}

func (e nilPidError) Error() string {
	return "nil_pid"
}

//
// IsChannelFullError checks if error is a ChannelFullError
//
func IsChannelFullError(err error) bool {
	_, ok := err.(chanFullError)
	return ok
}

func (e chanFullError) Error() string {
	return "chan_full"
}

//
// IsAlreadyRegError checks if error is an AlreadyRegError
//
func IsAlreadyRegError(err error) bool {
	_, ok := err.(alreadyRegError)
	return ok
}

func (e alreadyRegError) Error() string {
	return "already_registered"
}

//
// IsNotRegError checks if error is a NotRegError
//
func IsNotRegError(err error) bool {
	_, ok := err.(notRegError)
	return ok
}

func (e notRegError) Error() string {
	return "not_reg"
}

//
// IsNameEmptyError checks if error is a NameEmptyError
//
func IsNameEmptyError(err error) bool {
	_, ok := err.(nameEmptyError)
	return ok
}

func (e nameEmptyError) Error() string {
	return "name_empty"
}

//
// IsPrefixEmptyError checks if error is a PrefixEmptyError
//
func IsPrefixEmptyError(err error) bool {
	_, ok := err.(prefixEmptyError)
	return ok
}

func (e prefixEmptyError) Error() string {
	return "prefix_empty"
}

//
// IsExitNormalError checks if error is an ExitNormalError
//
func IsExitNormalError(e error) bool {
	return e.Error() == ExitNormal
}
