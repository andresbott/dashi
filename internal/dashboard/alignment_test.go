package dashboard

import "testing"

func TestPlaceContainer(t *testing.T) {
	tests := []struct {
		name       string
		horizontal string
		vertical   string
		want       ContainerPlacement
	}{
		{"top left", "left", "top", ContainerPlacement{MarginLeft: "0", MarginRight: "auto", JustifyContent: "flex-start"}},
		{"center center", "center", "center", ContainerPlacement{MarginLeft: "auto", MarginRight: "auto", JustifyContent: "center"}},
		{"bottom right", "right", "bottom", ContainerPlacement{MarginLeft: "auto", MarginRight: "0", JustifyContent: "flex-end"}},
		// Dashboards saved without a value (or with a hand-edited one) keep
		// the historical defaults of both server renders: centred, at the top.
		{"empty defaults", "", "", ContainerPlacement{MarginLeft: "auto", MarginRight: "auto", JustifyContent: "flex-start"}},
		{"unknown defaults", "middle", "sideways", ContainerPlacement{MarginLeft: "auto", MarginRight: "auto", JustifyContent: "flex-start"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PlaceContainer(tt.horizontal, tt.vertical); got != tt.want {
				t.Errorf("PlaceContainer(%q, %q) = %+v, want %+v", tt.horizontal, tt.vertical, got, tt.want)
			}
		})
	}
}
