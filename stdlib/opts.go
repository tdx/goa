package stdlib

//
// SpawnOpts for spawn process
//

//
// NewSpawnOpts makes spawn object and returns object to manipulate
//
func NewSpawnOpts() *SpawnOpts {
	return new(SpawnOpts)
}

//
// SpawnOpts is the structure to hold values of the options
//
type SpawnOpts struct {
	Prefix      string
	Name        Term
	UsrChanSize int
	SysChanSize int
	//
	sysOpts
}

type sysOpts struct {
	link                  bool
	linkPid               *Pid
	returnPidIfRegistered bool
	tracer                Tracer
}

//
// WithName sets name
//
func (op *SpawnOpts) WithName(name string) *SpawnOpts {

	op.Name = name

	return op
}

//
// WithPrefix sets prefix
//
func (op *SpawnOpts) WithPrefix(prefix string) *SpawnOpts {

	op.Prefix = prefix

	return op
}

//
// WithLinkTo sets pid to link to the process
//
func (op *SpawnOpts) WithLinkTo(pid *Pid) *SpawnOpts {

	op.link = true
	op.linkPid = pid

	return op
}

//
// WithUsrChannelSize sets size of the usr channel
//
func (op *SpawnOpts) WithUsrChannelSize(size int) *SpawnOpts {

	op.UsrChanSize = size

	return op
}

//
// WithSysChannelSize sets size of the sys channel
//
func (op *SpawnOpts) WithSysChannelSize(size int) *SpawnOpts {

	op.SysChanSize = size

	return op
}

//
// WithSpawnOrLocate sets flag to return process pid if the process with
//  given name and prefix already registered
//
func (op *SpawnOpts) WithSpawnOrLocate() *SpawnOpts {

	op.returnPidIfRegistered = true

	return op
}

//
// WithTracer sets tracer for the process
//
func (op *SpawnOpts) WithTracer(t Tracer) *SpawnOpts {

	op.tracer = t

	return op
}
