package flag

import flags "github.com/jessevdk/go-flags"

type ExistingOrgName string

var ExistingOrgNameCompleteFunc = func(prefix string) []flags.Completion {
	return []flags.Completion{}
}

func (ExistingOrgName) Complete(prefix string) []flags.Completion {
	return ExistingOrgNameCompleteFunc(prefix)
}

func (o ExistingOrgName) String() string {
	return string(o)
}
