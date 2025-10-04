package testhelpers

import (
	"flag"
	"runtime"
	"strings"
)

var testFlags = []string{
	"test.v", "test.run", "test.bench", "test.cpu",
	"test.timeout", "test.coverprofile", "test.covermode",
	"test.count", "test.parallel", "test.short", "test.failfast",
}

// IsTestEnv checks if the code is running in test mode
func IsTestEnv() bool {
	for _, flagName := range testFlags {
		if flag.Lookup(flagName) != nil {
			return true
		}
	}

	return isTestFunc()
}

func isTestFunc() bool {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return false
	}
	fn := runtime.FuncForPC(pc)
	return strings.Contains(fn.Name(), ".Test")
}
