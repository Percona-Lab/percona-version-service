package util

import (
	gover "github.com/hashicorp/go-version"
)

func Goversion(v string) *gover.Version {
	return gover.Must(gover.NewVersion(v))
}
