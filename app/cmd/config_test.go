package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizePublicURLs(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		want      string
		wantWarn  string // substring of the single expected warning; empty means none
		notInWarn string // must not appear in the warning
	}{
		{
			name: "trailing slash is stripped",
			raw:  "https://ha.example.com/hassio/ingress/dashi/",
			want: "https://ha.example.com/hassio/ingress/dashi",
		},
		{
			name: "encoded path is preserved",
			raw:  "https://ha.example.com/a%2Fb/my%20dash/",
			want: "https://ha.example.com/a%2Fb/my%20dash",
		},
		{
			name:     "query and fragment are dropped with a warning",
			raw:      "https://edit.example.com/base/?token=x#frag",
			want:     "https://edit.example.com/base",
			wantWarn: "dropped query, fragment or user info",
		},
		{
			name:      "userinfo is dropped with a warning that hides the credential",
			raw:       "https://admin:s3cret@edit.example.com/",
			want:      "https://edit.example.com",
			wantWarn:  "dropped query, fragment or user info",
			notInWarn: "s3cret",
		},
		{
			name:     "non http scheme is rejected",
			raw:      "ftp://edit.example.com",
			want:     "",
			wantWarn: "not an absolute http or https URL",
		},
		{
			name:      "rejected schemeless value with credentials hides the password",
			raw:       "admin:secret@ha.local:8123",
			want:      "",
			wantWarn:  "not an absolute http or https URL",
			notInWarn: "secret",
		},
		{
			name:      "rejected non http value with userinfo hides the password",
			raw:       "ftp://u:secret@edit.example.com",
			want:      "",
			wantWarn:  "not an absolute http or https URL",
			notInWarn: "secret",
		},
		{
			name:     "bare question mark is dropped with a warning",
			raw:      "https://edit.example.com/base?",
			want:     "https://edit.example.com/base",
			wantWarn: "dropped query, fragment or user info",
		},
		{
			name:     "fragment only is dropped with a warning",
			raw:      "https://edit.example.com/#frag",
			want:     "https://edit.example.com",
			wantWarn: "dropped query, fragment or user info",
		},
		{
			name:     "bare fragment value is rejected",
			raw:      "#frag",
			want:     "",
			wantWarn: "not an absolute http or https URL",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &AppCfg{}
			cfg.Server.Editor.PublicUrl = tc.raw
			normalizePublicURLs(cfg)

			if cfg.Server.Editor.PublicUrl != tc.want {
				t.Errorf("editor public url = %q, want %q", cfg.Server.Editor.PublicUrl, tc.want)
			}

			var warns []string
			for _, m := range cfg.Msgs {
				if m.Level == "warn" {
					warns = append(warns, m.Msg)
				}
			}
			if tc.wantWarn == "" {
				if len(warns) != 0 {
					t.Fatalf("warnings = %q, want none", warns)
				}
				return
			}
			if len(warns) != 1 {
				t.Fatalf("warnings = %q, want one mentioning %q", warns, tc.wantWarn)
			}
			if !strings.Contains(warns[0], tc.wantWarn) || !strings.Contains(warns[0], "server.editor.publicUrl") {
				t.Errorf("warning = %q, want it to mention server.editor.publicUrl and %q", warns[0], tc.wantWarn)
			}
			if tc.notInWarn != "" && strings.Contains(warns[0], tc.notInWarn) {
				t.Errorf("warning = %q, must not contain %q", warns[0], tc.notInWarn)
			}
		})
	}
}

func TestGetAppCfgPublicURLs(t *testing.T) {
	tests := []struct {
		name       string
		yaml       string
		env        map[string]string
		wantViewer string
		wantEditor string
		wantWarns  []string // substrings, one warning each, in order
	}{
		{
			name: "empty by default",
		},
		{
			name: "trailing slash is stripped",
			yaml: "Server:\n  Viewer:\n    PublicUrl: \"https://dash.example.com/\"\n" +
				"  Editor:\n    PublicUrl: \"https://ha.example.com/hassio/ingress/dashi/\"\n",
			wantViewer: "https://dash.example.com",
			wantEditor: "https://ha.example.com/hassio/ingress/dashi",
		},
		{
			name:       "surrounding whitespace is trimmed",
			yaml:       "Server:\n  Viewer:\n    PublicUrl: \"  http://10.0.0.5:8087  \"\n",
			wantViewer: "http://10.0.0.5:8087",
		},
		{
			name: "non http scheme is dropped with a warning",
			yaml: "Server:\n  Viewer:\n    PublicUrl: \"javascript:alert(1)\"\n" +
				"  Editor:\n    PublicUrl: \"ftp://edit.example.com\"\n",
			wantWarns: []string{"server.viewer.publicUrl", "server.editor.publicUrl"},
		},
		{
			name:      "url without host is dropped with a warning",
			yaml:      "Server:\n  Editor:\n    PublicUrl: \"/dashboards\"\n",
			wantWarns: []string{"server.editor.publicUrl"},
		},
		{
			name:       "editor url from environment",
			env:        map[string]string{"DASHI_SERVER_EDITOR_PUBLICURL": "https://edit.example.com/"},
			wantEditor: "https://edit.example.com",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Keep a developer's real DASHI_* settings out of the result.
			t.Setenv("DASHI_SERVER_VIEWER_PUBLICURL", "")
			t.Setenv("DASHI_SERVER_EDITOR_PUBLICURL", "")
			for k, v := range tc.env {
				t.Setenv(k, v)
			}
			file := filepath.Join(t.TempDir(), "config.yaml")
			if tc.yaml != "" {
				if err := os.WriteFile(file, []byte(tc.yaml), 0o600); err != nil {
					t.Fatalf("write config: %v", err)
				}
			}

			cfg, err := getAppCfg(file)
			if err != nil {
				t.Fatalf("getAppCfg: %v", err)
			}
			if cfg.Server.Viewer.PublicUrl != tc.wantViewer {
				t.Errorf("viewer public url = %q, want %q", cfg.Server.Viewer.PublicUrl, tc.wantViewer)
			}
			if cfg.Server.Editor.PublicUrl != tc.wantEditor {
				t.Errorf("editor public url = %q, want %q", cfg.Server.Editor.PublicUrl, tc.wantEditor)
			}

			var warns []string
			for _, m := range cfg.Msgs {
				if m.Level == "warn" {
					warns = append(warns, m.Msg)
				}
			}
			if len(warns) != len(tc.wantWarns) {
				t.Fatalf("warnings = %q, want %d matching %q", warns, len(tc.wantWarns), tc.wantWarns)
			}
			for i, want := range tc.wantWarns {
				if !strings.Contains(warns[i], want) {
					t.Errorf("warning %d = %q, want it to mention %q", i, warns[i], want)
				}
			}
		})
	}
}
