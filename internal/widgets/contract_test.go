package widgets

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

// TestConfigFieldsMatchContract pins the config JSON each widget's Go renderer
// reads. webui/src/widgets/contract.test.ts checks the editor's types against
// the same file, so a field added on one side only fails. It parses the source
// because the structs are unexported, which also finds new widgets by itself:
// any `type <key>Config struct` under a widget package must be in the contract.
func TestConfigFieldsMatchContract(t *testing.T) {
	raw, err := os.ReadFile("testdata/widget_contract.json")
	if err != nil {
		t.Fatalf("read contract: %v", err)
	}
	var contract struct {
		Configs map[string]struct {
			Fields []string `json:"fields"`
		} `json:"configs"`
	}
	if err := json.Unmarshal(raw, &contract); err != nil {
		t.Fatalf("parse contract: %v", err)
	}

	want := map[string][]string{}
	for key, cfg := range contract.Configs {
		if len(cfg.Fields) > 0 {
			want[key] = slices.Sorted(slices.Values(cfg.Fields))
		}
	}

	got := goConfigFields(t)

	gotKeys := slices.Sorted(maps.Keys(got))
	wantKeys := slices.Sorted(maps.Keys(want))
	if !slices.Equal(gotKeys, wantKeys) {
		t.Errorf("Go config structs differ from the contract\n got: %v\nwant: %v", gotKeys, wantKeys)
	}
	for key, fields := range want {
		t.Run(key, func(t *testing.T) {
			if !slices.Equal(got[key], fields) {
				t.Errorf("json fields differ from the contract\n got: %v\nwant: %v", got[key], fields)
			}
		})
	}
}

// goConfigFields returns the sorted JSON field names of every `<key>Config`
// struct in the widget packages, by key.
func goConfigFields(t *testing.T) map[string][]string {
	t.Helper()
	paths, err := filepath.Glob("*/*.go")
	if err != nil {
		t.Fatalf("glob widget sources: %v", err)
	}
	found := map[string][]string{}
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		ast.Inspect(file, func(n ast.Node) bool {
			spec, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}
			st, ok := spec.Type.(*ast.StructType)
			if ok && strings.HasSuffix(spec.Name.Name, "Config") {
				key := strings.TrimSuffix(spec.Name.Name, "Config")
				found[key] = jsonFieldNames(t, path, st)
			}
			return true
		})
	}
	return found
}

func jsonFieldNames(t *testing.T, path string, st *ast.StructType) []string {
	t.Helper()
	names := make([]string, 0, len(st.Fields.List))
	for _, field := range st.Fields.List {
		if field.Tag == nil {
			t.Errorf("%s: config field %v has no json tag", path, field.Names)
			continue
		}
		tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`")).Get("json")
		name, _, _ := strings.Cut(tag, ",")
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}
