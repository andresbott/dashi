package widgets

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/andresbott/dashi/internal/dashboard"
)

type fakeStore struct {
	list []dashboard.DashboardMeta
	dash map[string]dashboard.Dashboard
	// errors for specific ids
	getErr map[string]error
	// error from List
	listErr error
}

func (f *fakeStore) List() ([]dashboard.DashboardMeta, error) {
	return f.list, f.listErr
}

func (f *fakeStore) Get(id string) (dashboard.Dashboard, error) {
	if err, ok := f.getErr[id]; ok {
		return dashboard.Dashboard{}, err
	}
	return f.dash[id], nil
}

func TestCollectConfigs_EmptyStore(t *testing.T) {
	got := CollectConfigs(&fakeStore{}, "weather")
	if len(got) != 0 {
		t.Errorf("expected 0 configs, got %d", len(got))
	}
}

func TestCollectConfigs_FiltersByType(t *testing.T) {
	store := &fakeStore{
		list: []dashboard.DashboardMeta{{ID: "a"}},
		dash: map[string]dashboard.Dashboard{
			"a": {
				Pages: []dashboard.Page{{
					Rows: []dashboard.Row{{
						Widgets: []dashboard.Widget{
							{Type: "weather", Config: json.RawMessage(`{"lat":1}`)},
							{Type: "market", Config: json.RawMessage(`{"symbol":"X"}`)},
							{Type: "weather", Config: json.RawMessage(`{"lat":2}`)},
						},
					}},
				}},
			},
		},
	}

	got := CollectConfigs(store, "weather")
	if len(got) != 2 {
		t.Fatalf("expected 2 weather configs, got %d", len(got))
	}
	if string(got[0]) != `{"lat":1}` || string(got[1]) != `{"lat":2}` {
		t.Errorf("unexpected configs: %v", got)
	}
}

func TestCollectConfigs_NestedPagesAndRows(t *testing.T) {
	store := &fakeStore{
		list: []dashboard.DashboardMeta{{ID: "a"}},
		dash: map[string]dashboard.Dashboard{
			"a": {
				Pages: []dashboard.Page{
					{Rows: []dashboard.Row{{Widgets: []dashboard.Widget{{Type: "clock"}}}}},
					{Rows: []dashboard.Row{{Widgets: []dashboard.Widget{{Type: "clock"}}}}},
				},
			},
		},
	}
	got := CollectConfigs(store, "clock")
	if len(got) != 2 {
		t.Errorf("expected 2 clock configs across pages, got %d", len(got))
	}
}

func TestCollectConfigs_SkipsGetErrors(t *testing.T) {
	store := &fakeStore{
		list: []dashboard.DashboardMeta{{ID: "a"}, {ID: "b"}},
		dash: map[string]dashboard.Dashboard{
			"b": {Pages: []dashboard.Page{{Rows: []dashboard.Row{{Widgets: []dashboard.Widget{{Type: "clock"}}}}}}},
		},
		getErr: map[string]error{"a": errors.New("boom")},
	}
	got := CollectConfigs(store, "clock")
	if len(got) != 1 {
		t.Errorf("expected to skip errored dashboard and return 1, got %d", len(got))
	}
}

func TestCollectConfigs_ListError(t *testing.T) {
	store := &fakeStore{listErr: errors.New("boom")}
	got := CollectConfigs(store, "clock")
	if len(got) != 0 {
		t.Errorf("expected 0 on list error, got %d", len(got))
	}
}
