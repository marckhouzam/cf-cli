package flag

import flags "github.com/jessevdk/go-flags"

type ExistingOrgType string

var ExistingOrgTypeCompleteFunc = func(prefix string) []flags.Completion {
	return []flags.Completion{}
}

func (ExistingOrgType) Complete(prefix string) []flags.Completion {
	return ExistingOrgTypeCompleteFunc(prefix)
}

func (o ExistingOrgType) String() string {
	return string(o)
}
