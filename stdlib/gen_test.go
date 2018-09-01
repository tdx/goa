package stdlib

import (
	"fmt"
	"testing"
	"time"
)

func TestGenSendNilPid(t *testing.T) {
	var pid *Pid

	if err := pid.Send("test"); !IsNilPidError(err) {
		t.Fatalf("expected %s error, actual %s", NilPidError, err)
	}

	if _, err := pid.Call("test"); !IsNilPidError(err) {
		t.Fatalf("expected %s error, actual %s", NilPidError, err)
	}
}

func TestGenChannelErrorFull(t *testing.T) {
	pid, err := SpawnWithOpts(
		genTestFailFunc,
		NewSpawnOpts().WithUsrChannelSize(1))

	if err = pid.Send("test1"); err != nil {
		t.Fatal(err)
	}
	if err = pid.Send("test2"); !IsChannelFullError(err) {
		t.Fatalf("expected %s error, actual %s", ChannelFullError, err)
	}
	if _, err = pid.Call("test3"); !IsChannelFullError(err) {
		t.Fatalf("expected %s error, actual %s", ChannelFullError, err)
	}

	pid2, err := SpawnWithOpts(
		genTestFailFunc,
		NewSpawnOpts().WithSysChannelSize(1))

	if err = pid2.SendSys("test1"); err != nil {
		t.Fatal(err)
	}
	if err = pid2.SendSys("test2"); !IsChannelFullError(err) {
		t.Fatalf("expected %s error, actual %s", ChannelFullError, err)
	}
	if _, err = pid2.CallSys("test3"); !IsChannelFullError(err) {
		t.Fatalf("expected %s error, actual %s", ChannelFullError, err)
	}
}

func TestGenChannelErrorNoProc(t *testing.T) {
	pid, err := Spawn(genTestFastReturnFunc)

	time.Sleep(time.Duration(50) * time.Millisecond)

	if err = pid.Send("test1"); !IsNoProcError(err) {
		t.Fatalf("expected %s error, actual %s", NoProcError, err)
	}
	if _, err = pid.Call("test2"); !IsNoProcError(err) {
		t.Fatalf("expected %s error, actual %s", NoProcError, err)
	}

	pid2, err := Spawn(genTestFastReturnFunc)

	time.Sleep(time.Duration(50) * time.Millisecond)

	if err = pid2.SendSys("test1"); !IsNoProcError(err) {
		t.Fatalf("expected %s error, actual %s", NoProcError, err)
	}
	if _, err = pid2.CallSys("test2"); !IsNoProcError(err) {
		t.Fatalf("expected %s error, actual %s", NoProcError, err)
	}
}

//
// Locals
//
func genTestFailFunc(gp GenProc, args ...Term) error {
	time.Sleep(time.Duration(100) * time.Millisecond)
	return fmt.Errorf("func failed")
}

func genTestFastReturnFunc(gp GenProc, args ...Term) error {
	return nil
}
