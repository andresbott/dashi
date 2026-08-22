package spa

import (
	"embed"
	"io/fs"
	"net/http"

	handlers "github.com/go-bumbu/http/handlers/spa"
)

//go:embed all:files/ui
var UiFiles embed.FS

func App(path string) (http.Handler, error) {
	return handlers.NewSpaHAndler(
		UiFiles,
		"files/ui",
		path,
	)
}

// IndexHTML returns the embedded SPA shell (index.html).
func IndexHTML() ([]byte, error) {
	return fs.ReadFile(UiFiles, "files/ui/index.html")
}

// FileExists reports whether name (relative to the SPA root, no leading slash)
// is a real embedded file — used to serve assets/favicons verbatim while every
// other path falls back to the injected shell.
func FileExists(name string) bool {
	f, err := UiFiles.Open("files/ui/" + name)
	if err != nil {
		return false
	}
	defer func() { _ = f.Close() }()
	st, err := f.Stat()
	return err == nil && !st.IsDir()
}
