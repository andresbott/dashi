package router

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

// TestWidgetTypesMatchContract pins the widget types the backend registers.
// webui/src/lib/widgetRegistry.test.ts checks the same file for the frontend,
// so a widget added on one side only fails. It lives here because only
// newSharedDeps holds the full module list.
func TestWidgetTypesMatchContract(t *testing.T) {
	raw, err := os.ReadFile("../../internal/widgets/testdata/widget_contract.json")
	if err != nil {
		t.Fatalf("read contract: %v", err)
	}
	var contract struct {
		Types []string `json:"types"`
	}
	if err := json.Unmarshal(raw, &contract); err != nil {
		t.Fatalf("parse contract: %v", err)
	}

	newTestViewerWithDashboards(t) // builds sharedTestDeps

	got := make([]string, 0, len(sharedTestDeps.modules))
	for _, m := range sharedTestDeps.modules {
		got = append(got, m.Type())
	}
	slices.Sort(got)
	want := slices.Sorted(slices.Values(contract.Types))
	if !slices.Equal(got, want) {
		t.Fatalf("registered widget types differ from the contract\n got: %v\nwant: %v", got, want)
	}
}
