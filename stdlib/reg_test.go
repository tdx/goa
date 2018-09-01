package stdlib

import (
	"testing"
	"time"
)

const (
	regTestName      = "regTest#12345"
	regTestName2     = "regTest2"
	regTestName3     = "regTest3"
	regTestPrefixOne = "groupOne"
	regTestPrefixTwo = "groupTwo"
)

func TestGenServerRegName(t *testing.T) {
	var nilPid *Pid

	err := nilPid.Register("")
	if err == nil {
		t.Fatalf("expected '%s' error, actual no error", NilPidError)
	}
	if err != nil && !IsNilPidError(err) {
		t.Fatalf("expected '%s' error, actual %s'", NilPidError, err)
	}

	pid, err := GenServerStart(new(rs))
	if err != nil {
		t.Fatal(err)
	}

	err = pid.Register("")
	if !IsNameEmptyError(err) {
		t.Fatalf("exprected '%s' error, actual '%s'", NameEmptyError, err)
	}

	err = pid.Register(regTestName)
	if err != nil {
		t.Fatal(err)
	}

	//
	// reg with same name is AlreadyRegError
	//
	err = pid.Register(regTestName)
	if !IsAlreadyRegError(err) {
		t.Fatalf("exprected '%s' error, actual '%s'", AlreadyRegError, err)
	}

	pid2, err := GenServerStart(new(rs))
	if err != nil {
		t.Fatal(err)
	}

	//
	// try with new pid
	//
	err = pid2.Register(regTestName)
	if !IsAlreadyRegError(err) {
		t.Fatalf("exprected '%s' error, actual '%s'", AlreadyRegError, err)
	}

	err = pid2.Register(regTestName3)
	if err != nil {
		t.Fatal(err)
	}

	//
	// but diff name is ok
	//
	err = pid.Register(regTestName2)
	if err != nil {
		t.Fatal(err)
	}

	err = pid.Stop()
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Duration(300) * time.Millisecond)

	//
	// name is unregistered when process exits
	//

	//
	// check name is available to register again
	//
	pid, err = GenServerStart(new(rs))
	if err != nil {
		t.Fatal(err)
	}
	err = pid.Register(regTestName)
	if err != nil {
		t.Fatal(err)
	}
	_ = pid.Stop()
	_ = pid2.Stop()
}

func TestGenServerRegUnregName(t *testing.T) {

	var nilPid *Pid

	err := nilPid.Unregister("")
	if !IsNilPidError(err) {
		t.Fatalf("expected '%s' error, actual %s'", NilPidError, err)
	}

	pid, err := GenServerStart(new(rs))
	if err != nil {
		t.Fatal(err)
	}

	err = pid.Unregister("")
	if !IsNameEmptyError(err) {
		t.Fatalf("exprected '%s' error, actual '%s'", NameEmptyError, err)
	}

	//
	// Unregister empty or not registered name - error
	//
	err = pid.Unregister("")
	if !IsNameEmptyError(err) {
		t.Fatalf("exprected '%s' error, actual '%s'", NameEmptyError, err)
	}

	err = pid.Unregister(regTestName)
	if !IsNotRegError(err) {
		t.Fatalf("exprected '%s' error, actual '%s'", NotRegError, err)
	}

	//
	// reg/unreg normal is ok
	//
	err = pid.Register(regTestName)
	if err != nil {
		t.Fatal(err)
	}

	err = pid.Unregister(regTestName)
	if err != nil {
		t.Fatal(err)
	}

	// 2 name reg for pid
	if err := pid.Register(regTestName); err != nil {
		t.Fatal(err)
	}
	if err := pid.Register(regTestName2); err != nil {
		t.Fatal(err)
	}
	if err := pid.Unregister(regTestName); err != nil {
		t.Fatal(err)
	}
	if err := pid.Unregister(regTestName2); err != nil {
		t.Fatal(err)
	}

	// no regs in new env
	env2 := NewEnv()
	pid2, err := env2.GenServerStart(new(rs))
	err = pid2.Unregister(regTestName)
	if !IsNotRegError(err) {
		t.Fatalf("exprected '%s' error, actual '%s'", NameEmptyError, err)
	}
	err = pid2.UnregisterPrefix(regTestPrefixOne, regTestName)
	if !IsNotRegError(err) {
		t.Fatalf("exprected '%s' error, actual '%s'", NameEmptyError, err)
	}
	if err := pid2.RegisterPrefix(regTestPrefixTwo, regTestName2); err != nil {
		t.Fatal(err)
	}
	if err := pid2.Stop(); err != nil {
		t.Fatal(err)
	}
}

func TestGenServerRegWhereis(t *testing.T) {

	env2 := NewEnv()
	_, err := env2.Whereis("")
	if !IsNameEmptyError(err) {
		t.Fatalf("expected '%s' error, actual %s'", NameEmptyError, err)
	}

	pid, err := env2.Whereis(regTestName)
	if !IsNotRegError(err) {
		t.Fatalf("expected '%s' error, actual '%s'", NotRegError, err)
	}
	if pid != nil {
		t.Fatalf("expected no pid has name '%s', actual %s", regTestName, pid)
	}

	pid, err = env2.GenServerStart(new(rs))
	if err != nil {
		t.Fatal(err)
	}
	err = pid.Register(regTestName3)
	if err != nil {
		t.Fatal(err)
	}

	pid2, err := env2.Whereis(regTestName3)
	if err != nil {
		t.Fatal(err)
	}
	if !pid2.Equal(pid) {
		t.Fatalf("expected name '%s' was register for pid %s, actual for %s",
			regTestName3, pid, pid2)
	}

	err = pid.Unregister(regTestName3)
	if err != nil {
		t.Fatal(err)
	}

	pid3, err := env2.Whereis(regTestName3)
	if !IsNotRegError(err) {
		t.Fatalf("expected '%s' error, actual '%s'", NotRegError, err)
	}
	if pid3 != nil {
		t.Fatalf("expected no pids with name '%s', actual %s",
			regTestName3, pid3)
	}
}

func TestGenServerRegPrefixName(t *testing.T) {
	var nilPid *Pid

	err := nilPid.RegisterPrefix("", regTestName)
	if err == nil {
		t.Fatalf("expected '%s' error, actual no error", NilPidError)
	}
	if err != nil && !IsNilPidError(err) {
		t.Fatalf("expected '%s' error, actual %s'", NilPidError, err)
	}

	pid, err := GenServerStart(new(ts))
	if err != nil {
		t.Fatal(err)
	}

	err = pid.RegisterPrefix("", regTestName)
	if !IsPrefixEmptyError(err) {
		t.Fatalf("exprected '%s' error, actual '%s'", PrefixEmptyError, err)
	}

	err = pid.RegisterPrefix("", "")
	if !IsPrefixEmptyError(err) {
		t.Fatalf("exprected '%s' error, actual '%s'", PrefixEmptyError, err)
	}

	err = pid.RegisterPrefix(regTestPrefixOne, "")
	if !IsNameEmptyError(err) {
		t.Fatalf("exprected '%s' error, actual '%s'", NameEmptyError, err)
	}

	err = pid.RegisterPrefix(regTestPrefixOne, regTestName)
	if err != nil {
		t.Fatal(err)
	}

	err = pid.RegisterPrefix(regTestPrefixOne, regTestName)
	if !IsAlreadyRegError(err) {
		t.Fatalf("expected '%s' error, actual %s'", AlreadyRegError, err)
	}

	err = pid.UnregisterPrefix(regTestPrefixOne, regTestName)
	if err != nil {
		t.Fatal(err)
	}

	err = pid.UnregisterPrefix(regTestPrefixOne, regTestName)
	if !IsNotRegError(err) {
		t.Fatalf("expected '%s' error, actual %s'", NotRegError, err)
	}

	env2 := NewEnv()
	pid2, err := env2.GenServerStart(new(rs))
	if err != nil {
		t.Fatal(err)
	}

	if err := pid2.RegisterPrefix(regTestPrefixTwo, regTestName); err != nil {
		t.Fatal(err)
	}
}

func TestGenServerRegWhereare(t *testing.T) {

	env2 := NewEnv()
	_, err := env2.Whereare("")
	if !IsPrefixEmptyError(err) {
		t.Fatalf("expected '%s' error, actual %s'", PrefixEmptyError, err)
	}

	pids, err := env2.Whereare(regTestName)
	if !IsNotRegError(err) {
		t.Fatalf("expected '%s' error, actual '%s'", NotRegError, err)
	}

	pid, err := env2.GenServerStart(new(rs))
	if err != nil {
		t.Fatal(err)
	}
	err = pid.RegisterPrefix(regTestPrefixOne, regTestName)
	if err != nil {
		t.Fatal(err)
	}

	pids, err = env2.Whereare(regTestPrefixTwo)
	if !IsNotRegError(err) {
		t.Fatalf("expected '%s' error, actual '%s'", NotRegError, err)
	}

	pids, err = env2.Whereare(regTestPrefixOne)
	if err != nil {
		t.Fatal(err)
	}
	if len(pids) != 1 {
		t.Fatalf("expected group '%s' has 1 member, actual %d",
			regTestPrefixOne, len(pids))
	}

	err = pid.UnregisterPrefix(regTestPrefixOne, regTestName)
	if err != nil {
		t.Fatal(err)
	}

	pids, err = env2.Whereare(regTestPrefixOne)
	if len(pids) != 0 {
		t.Fatalf("expected group '%s' has 0 member, actual %d",
			regTestPrefixOne, len(pids))
	}

	err = pid.RegisterPrefix(regTestPrefixOne, regTestName)
	if err != nil {
		t.Fatal(err)
	}
	err = pid.RegisterPrefix(regTestPrefixOne, regTestName2)
	if err != nil {
		t.Fatal(err)
	}
	err = pid.RegisterPrefix(regTestPrefixOne, regTestName3)
	if err != nil {
		t.Fatal(err)
	}

	pids, err = env2.Whereare(regTestPrefixOne)
	if len(pids) != 3 {
		t.Fatalf("expected group '%s' has 3 member, actual %d",
			regTestPrefixOne, len(pids))
	}
}

//
// GenServet to test registrations
//
type rs struct {
	GenServerSys
}
