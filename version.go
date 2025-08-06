// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package wxx

import (
	"github.com/maloquacious/semver"
)

func Version() semver.Version {
	return semver.Version{
		Major:      0,
		Minor:      15,
		Patch:      0,
		PreRelease: "alpha",
		Build:      semver.Commit(),
	}
}
