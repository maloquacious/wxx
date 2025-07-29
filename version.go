// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package wxx

import (
	"github.com/maloquacious/semver"
)

func Version() semver.Version {
	return semver.Version{
		Major:      0,
		Minor:      2,
		Patch:      1,
		PreRelease: "alpha",
	}
}
