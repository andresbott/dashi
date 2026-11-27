package cmd

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-bumbu/config"
)

type AppCfg struct {
	Server  serverCfg
	Obs     obsCfg `config:"Observability"`
	Auth    authConfig
	Env     Env
	Msgs    []Msg
	DataDir string
}

type Env struct {
	LogLevel   string
	Production bool
}

type serverCfg struct {
	Viewer viewerCfg
	Editor editorCfg
}

type viewerCfg struct {
	Enabled bool
	BindIp  string
	Port    int
	// PublicUrl is the URL browsers use to reach the viewer, e.g.
	// "https://dash.example.com". Only needed when the derived
	// "<editor host>:<viewer port>" is wrong — typically behind a reverse
	// proxy. The editor SPA reads it from GET /api/v0/info.
	PublicUrl string
}

func (c viewerCfg) Addr() string {
	if c.BindIp == "" {
		return ":" + strconv.Itoa(c.Port)
	}
	return c.BindIp + ":" + strconv.Itoa(c.Port)
}

type editorCfg struct {
	Enabled bool
	BindIp  string
	Port    int
	// PublicUrl is the URL browsers use to reach the editor, e.g.
	// "https://ha.example.com/hassio/ingress/dashi". It is never derived:
	// the viewer's Edit button only appears when it is set.
	PublicUrl string
}

func (c editorCfg) Addr() string {
	if c.BindIp == "" {
		return ":" + strconv.Itoa(c.Port)
	}
	return c.BindIp + ":" + strconv.Itoa(c.Port)
}

type obsCfg struct {
	Enabled bool
	BindIp  string
	Port    int
}

func (c obsCfg) Addr() string {
	if c.BindIp == "" {
		return ":" + strconv.Itoa(c.Port)
	}
	return c.BindIp + ":" + strconv.Itoa(c.Port)
}

type authConfig struct {
	Enabled     bool
	DefaultUser string `config:"DefaultUser"`
}

var defaultCfg = AppCfg{
	DataDir: "./data",
	Server: serverCfg{
		Viewer: viewerCfg{
			Enabled:   true,
			BindIp:    "",
			Port:      8087,
			PublicUrl: "",
		},
		Editor: editorCfg{
			Enabled:   true,
			BindIp:    "",
			Port:      8088,
			PublicUrl: "",
		},
	},
	Obs: obsCfg{
		Enabled: false,
		BindIp:  "",
		Port:    9090,
	},
	Auth: authConfig{
		Enabled:     false,
		DefaultUser: "default",
	},
	Env: Env{
		LogLevel:   "info",
		Production: false,
	},
}

type Msg struct {
	Level string
	Msg   string
}

const EnvBarPrefix = "DASHI"

func getAppCfg(file string) (AppCfg, error) {
	configMsg := []Msg{}
	cfg := AppCfg{}
	var err error
	_, err = config.Load(
		config.Defaults{Item: defaultCfg},
		config.EnvFile{Path: ".env", Mandatory: false},
		config.CfgFile{Path: file, Mandatory: false},
		config.EnvVar{Prefix: EnvBarPrefix},
		config.Unmarshal{Item: &cfg},
		config.Writer{Fn: func(level, msg string) {
			if level == config.InfoLevel {
				configMsg = append(configMsg, Msg{Level: "info", Msg: msg})
			}
			if level == config.DebugLevel {
				configMsg = append(configMsg, Msg{Level: "debug", Msg: msg})
			}
		}},
	)
	cfg.Msgs = configMsg
	if err != nil {
		return cfg, err
	}

	normalizePublicURLs(&cfg)

	absPath, err := filepath.Abs(cfg.DataDir)
	if err != nil {
		return cfg, fmt.Errorf("failed to get absolute path: %w", err)
	}
	cfg.DataDir = absPath

	if !cfg.Server.Viewer.Enabled && !cfg.Server.Editor.Enabled {
		return cfg, fmt.Errorf("at least one of server.viewer or server.editor must be enabled")
	}

	if !cfg.Auth.Enabled && cfg.Auth.DefaultUser == "" {
		cfg.Auth.DefaultUser = "default"
	}

	return cfg, nil
}

// normalizePublicURLs normalizes the configured public base URLs once, so
// every consumer can use them verbatim. A value that is not an absolute
// http/https URL is dropped with a warning: these URLs end up in hrefs, and a
// silent drop would leave the user guessing why a link is missing.
func normalizePublicURLs(cfg *AppCfg) {
	for _, u := range []struct {
		key string
		val *string
	}{
		{"server.viewer.publicUrl", &cfg.Server.Viewer.PublicUrl},
		{"server.editor.publicUrl", &cfg.Server.Editor.PublicUrl},
	} {
		raw := *u.val
		*u.val = normalizeBaseURL(raw)
		if *u.val == "" && strings.TrimSpace(raw) != "" {
			cfg.Msgs = append(cfg.Msgs, Msg{
				Level: "warn",
				Msg:   fmt.Sprintf("ignoring %s %q: not an absolute http or https URL", u.key, raw),
			})
		}
	}
}

// normalizeBaseURL validates a public base URL and strips its trailing slash.
// Anything that is not an absolute http/https URL comes back empty.
func normalizeBaseURL(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" {
		return ""
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return ""
	}
	return strings.TrimSuffix(u.Scheme+"://"+u.Host+u.Path, "/")
}
