package dashboard

import (
	"reflect"
	"testing"
)

func TestPlaceRow(t *testing.T) {
	tests := []struct {
		name    string
		widgets []Widget
		want    []Placement
	}{
		{
			name:    "empty row",
			widgets: nil,
			want:    []Placement{},
		},
		{
			name:    "no columns flow left to right",
			widgets: []Widget{{Width: 6}, {Width: 6}},
			want:    []Placement{{Gap: 0, Width: 6}, {Gap: 0, Width: 6}},
		},
		{
			name:    "width below one defaults to full span",
			widgets: []Widget{{Width: 0}},
			want:    []Placement{{Gap: 0, Width: 12}},
		},
		{
			name:    "explicit column produces a leading gap",
			widgets: []Widget{{Width: 6, Column: 4}},
			want:    []Placement{{Gap: 3, Width: 6}},
		},
		{
			name:    "column one is flush left with no gap",
			widgets: []Widget{{Width: 4, Column: 1}},
			want:    []Placement{{Gap: 0, Width: 4}},
		},
		{
			name:    "second widget honours its column relative to the cursor",
			widgets: []Widget{{Width: 3}, {Width: 6, Column: 7}},
			want:    []Placement{{Gap: 0, Width: 3}, {Gap: 3, Width: 6}},
		},
		{
			name:    "overlapping column is pushed right of the previous widget",
			widgets: []Widget{{Width: 6}, {Width: 6, Column: 3}},
			want:    []Placement{{Gap: 0, Width: 6}, {Gap: 0, Width: 6}},
		},
		{
			name:    "column beyond the grid is clamped to the last column",
			widgets: []Widget{{Width: 4, Column: 20}},
			want:    []Placement{{Gap: 11, Width: 4}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PlaceRow(tt.widgets)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PlaceRow(%+v) = %+v, want %+v", tt.widgets, got, tt.want)
			}
		})
	}
}
