package widgets

import (
	"encoding/json"

	"github.com/andresbott/dashi/internal/dashboard"
)

// DashboardLister is the minimal subset of *dashboard.Store that
// CollectConfigs needs. Defined as an interface so tests can use a fake
// and so internal/widgets does not gain a hard dependency on the Store
// implementation.
type DashboardLister interface {
	List() ([]dashboard.DashboardMeta, error)
	Get(id string) (dashboard.Dashboard, error)
}

// CollectConfigs walks every dashboard in the store, collecting the raw
// config JSON of every widget whose Type equals widgetType. Dashboards
// that fail to load are skipped silently — warmup is best-effort.
func CollectConfigs(store DashboardLister, widgetType string) []json.RawMessage {
	var out []json.RawMessage
	list, err := store.List()
	if err != nil {
		return out
	}
	for _, meta := range list {
		dash, err := store.Get(meta.ID)
		if err != nil {
			continue
		}
		for _, page := range dash.Pages {
			for _, row := range page.Rows {
				for _, w := range row.Widgets {
					if w.Type == widgetType {
						out = append(out, w.Config)
					}
				}
			}
		}
	}
	return out
}
