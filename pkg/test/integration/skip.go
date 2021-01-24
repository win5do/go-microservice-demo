package integration_test

import "os"

// SkipInCi skip integration test in CI
func SkipInCi() bool {
	return os.Getenv("CI") != ""
}
