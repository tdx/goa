package stdlib

import (
	// "fmt"
	"testing"
)

func TestRef(t *testing.T) {
	env1Ref1 := MakeRef()
	env1Ref2 := MakeRef()

	if env1Ref1.String() == env1Ref2.String() {
		t.Fatalf("expected %s != %s", env1Ref1, env1Ref2)
	}

	env2 := NewEnv()
	env2Ref1 := env2.MakeRef()
	env2Ref2 := env2.MakeRef()
	if env2Ref1.String() == env2Ref2.String() {
		t.Fatalf("expected %s != %s", env2Ref1, env2Ref2)
	}

}
