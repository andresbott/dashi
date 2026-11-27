package router

import (
	"testing"

	"github.com/andresbott/dashi/app/router/handlers"
)

func TestPublicEditor(t *testing.T) {
	tests := []struct {
		name string
		cfg  Cfg
		want handlers.PublicEditor
	}{
		{
			name: "enabled editor keeps its url",
			cfg:  Cfg{EditorEnabled: true, EditorPublicURL: "https://edit.example.com"},
			want: handlers.PublicEditor{Enabled: true, URL: "https://edit.example.com"},
		},
		{
			name: "enabled editor without url",
			cfg:  Cfg{EditorEnabled: true},
			want: handlers.PublicEditor{Enabled: true},
		},
		{
			name: "disabled editor drops its url",
			cfg:  Cfg{EditorEnabled: false, EditorPublicURL: "https://edit.example.com"},
			want: handlers.PublicEditor{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := publicEditor(tc.cfg); got != tc.want {
				t.Errorf("publicEditor() = %+v, want %+v", got, tc.want)
			}
		})
	}
}
