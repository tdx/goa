package stdlib

//
// Register associates the name with process
//
func (pid *Pid) Register(name Term) error {
	if pid == nil {
		return NilPidError
	}
	return pid.env.regPidName("", name, pid)
}

//
// Unregister removes process association with name
//
func (pid *Pid) Unregister(name Term) error {
	if pid == nil {
		return NilPidError
	}
	return pid.env.unregPidName("", name)
}

//
// Whereis lookups process by name in default environment
//
func Whereis(name Term) (*Pid, error) {
	return env.Whereis(name)
}

//
// Whereis lookups process by name in specified environment
//
func (e *Env) Whereis(name Term) (*Pid, error) {
	return e.whereis(name)
}

//
// RegisterPrefix registers name for process in the group of processes with
//  same prefix
//
func (pid *Pid) RegisterPrefix(prefix string, name Term) error {
	if pid == nil {
		return NilPidError
	}
	if prefix == "" {
		return PrefixEmptyError
	}
	return pid.env.regPidName(prefix, name, pid)
}

//
// UnregisterPrefix removes prefix and name association of the process
//
func (pid *Pid) UnregisterPrefix(prefix string, name Term) error {
	if pid == nil {
		return NilPidError
	}
	if prefix == "" {
		return PrefixEmptyError
	}
	return pid.env.unregPidName(prefix, name)
}

//
// WhereisPrefix lookups process by prefix and name in default environment
//
func WhereisPrefix(prefix string, name Term) (*Pid, error) {
	return env.WhereisPrefix(prefix, name)
}

//
// WhereisPrefix lookups process by prefix and name in specified environment
//
func (e *Env) WhereisPrefix(prefix string, name Term) (*Pid, error) {
	return e.whereisPrefix(prefix, name)
}

//
// Whereare lookups processes by prefix in default environment
//
func Whereare(prefix string) (RegMap, error) {
	return env.Whereare(prefix)
}

//
// Whereare lookups processes by prefix in specified environment
//
func (e *Env) Whereare(prefix string) (RegMap, error) {
	return e.whereare(prefix)
}
