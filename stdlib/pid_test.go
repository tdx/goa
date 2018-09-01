package stdlib

import (
	"testing"
)

func TestNilPids(t *testing.T) {
	var pid1, pid2 *Pid

	if pid1.ID() != 0 {
		t.Fatal("id of the nil pid must be 0")
	}

	if pid1.String() != "<nil>" {
		t.Fatal("String() of the nil pid must be <nil>")
	}

	if !pid1.Equal(pid2) {
		t.Fatal("nil values must be equal")
	}

	if ch := pid1.GetUsrChannel(); ch != nil {
		t.Fatal("user channel of the nil pid must bi nil")
	}

	if ch := pid1.GetSysChannel(); ch != nil {
		t.Fatal("system channel of the nil pid must bi nil")
	}
}
