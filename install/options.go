package install

// Options for installation.
type Options struct {
	SupervisorPath string
}

func (o Options) valid() bool {
	return o.SupervisorPath != ""
}
