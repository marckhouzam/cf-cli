package flag

import (
	flags "github.com/jessevdk/go-flags"
)

type ExistingAppName string

var ExistingAppNameCompleteFunc = func(prefix string) []flags.Completion { return nil }

func (ExistingAppName) Complete(prefix string) []flags.Completion {
	return ExistingAppNameCompleteFunc(prefix)
}
