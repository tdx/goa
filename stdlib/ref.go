package stdlib

import (
	"fmt"
)

//
// Ref is a unique reference
//
type Ref struct {
	envID uint32
	id    uint64
}

//
// MakeRef returns new ref from default environment
//
func MakeRef() Ref {
	return env.MakeRef()
}

//
// MakeRef returns new ref from specified environment
//
func (e *Env) MakeRef() Ref {
	return e.makeRef()
}

//
// String returns string presentaion of ref
//
func (r Ref) String() string {
	return fmt.Sprintf("#ref<0.%d.%d>", r.envID, r.id)
}

//
// CompareRefs compare 2 Refs
//
func CompareRefs(a, b Ref) int {
	switch {
	case a.envID > b.envID:
		return 1
	case a.envID < b.envID:
		return -1
	default:
		switch {
		case a.id > b.id:
			return 1
		case a.id < b.id:
			return -1
		default:
			return 0
		}
	}
}
