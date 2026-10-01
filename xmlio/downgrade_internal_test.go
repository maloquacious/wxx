// Copyright (c) 2026 Michael D Henderson. All rights reserved.

// This is an INTERNAL test (package xmlio, not xmlio_test), for the reason
// xml_header_internal_test.go gives: downgradeLoss is unexported, and the arm it
// is tested for here -- a schema with no loss inventory -- cannot be reached
// through the public API while every registered codec has one.
package xmlio

import (
	"errors"
	"strings"
	"testing"

	"github.com/maloquacious/wxx"
)

// TestDowngradeLossRefusesAnUninventoriedSchema holds downgradeLoss to the rule
// that losslessness is claimed per schema, never by default. A schema it has no
// arm for must be an error: answering "no loss" for it would let a newly
// registered codec report every encode as lossless without anyone having checked.
//
// "" is the case with history. It is the implicit legacy (classic) schema, whose
// inventory went with the classic codec (issue #103); if it ever came back it
// would need that inventory back, not a silent pass.
func TestDowngradeLossRefusesAnUninventoriedSchema(t *testing.T) {
	m := &wxx.Map_t{}

	// Positive control: every schema the registry writes has an arm and reports
	// no loss. Without it, an implementation that refused everything would pass
	// the cases below.
	for _, c := range codecs() {
		schema := c.AcceptedApps().Schema
		dropped, err := downgradeLoss(m, schema)
		if err != nil {
			t.Fatalf("downgradeLoss(schema %q): %v; every registered codec's schema must have a loss inventory", schema, err)
		}
		if len(dropped) != 0 {
			t.Errorf("downgradeLoss(schema %q) reported %d losses for an empty map, want none", schema, len(dropped))
		}
	}

	for _, schema := range []string{"", "1.07", "9.99", "garbage"} {
		dropped, err := downgradeLoss(m, schema)
		if err == nil {
			t.Errorf("downgradeLoss(schema %q) = %d losses, nil; want an error: no arm inventories this schema, so it cannot claim the target is lossless", schema, len(dropped))
			continue
		}
		if !errors.Is(err, wxx.ErrUnsupportedSchemaVersion) {
			t.Errorf("downgradeLoss(schema %q) error = %v, want it to wrap %v", schema, err, wxx.ErrUnsupportedSchemaVersion)
		}
		if !strings.Contains(err.Error(), `"`+schema+`"`) {
			t.Errorf("downgradeLoss(schema %q) error = %q, want it to name the schema", schema, err)
		}
		if dropped != nil {
			t.Errorf("downgradeLoss(schema %q) returned %d losses alongside its error, want none", schema, len(dropped))
		}
	}
}
