package flag

import flags "github.com/jessevdk/go-flags"

type AppType string

var AppTypeCompleteFunc = func(prefix string) []flags.Completion {
	return completions([]string{"buildpack", "docker"}, prefix, false)
}

func (AppType) Complete(prefix string) []flags.Completion {
	return AppTypeCompleteFunc(prefix)
}
