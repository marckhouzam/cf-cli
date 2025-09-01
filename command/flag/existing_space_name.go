package flag

import flags "github.com/jessevdk/go-flags"

type ExistingSpaceName string

var ExistingSpaceNameCompleteFunc = func(prefix string) []flags.Completion {
	return []flags.Completion{}
}

func (ExistingSpaceName) Complete(prefix string) []flags.Completion {
	return ExistingSpaceNameCompleteFunc(prefix)
}

func (o ExistingSpaceName) String() string {
	return string(o)
}
