package stdlib

import (
	// "fmt"
	"fmt"
	"testing"
	"time"
)

var npid int

func TestEnvDefault(t *testing.T) {
	if env == nil {
		t.Fatal("default env is nil")
	}
}

func TestEnvSpawn(t *testing.T) {

	if pid, err := env.Spawn(envTestFunc); err != nil {
		t.Fatal(err)
	} else if pid == nil {
		t.Fatalf("pid is nil")
	}

	if env2 := NewEnv(); env2 == nil {
		t.Fatalf("env is nil")
	} else if pid2, err := env2.Spawn(envTestFunc); err != nil {
		t.Fatal(err)
	} else if pid2 == nil {
		t.Fatalf("pid2 is nil")
	}

	time.Sleep(time.Duration(100) * time.Millisecond)
}

func TestEnvSpawnEmptyPid(t *testing.T) {
	var pid *Pid
	if _, err := pid.SpawnLink(envTestFunc); !IsNilPidError(err) {
		t.Fatalf("expected %s error, actual %s", NilPidError, err.Error())
	}
}

func BenchmarkNewPid(b *testing.B) {

	opts := NewSpawnOpts()

	if _, _, err := env.newPid(opts); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, _, err := env.newPid(opts); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNewPidWithName(b *testing.B) {

	opts := NewSpawnOpts()

	if _, _, err := env.newPid(opts); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		opts = opts.WithName(fmt.Sprintf("gs%d", npid))
		if _, _, err := env.newPid(opts); err != nil {
			b.Fatal(err)
		}
		npid++
	}
}

//
// Locals
//
func envTestFunc(gp GenProc, args ...Term) error {
	// fmt.Printf("new process with pid %s\n", gp.Self())
	return nil
}
