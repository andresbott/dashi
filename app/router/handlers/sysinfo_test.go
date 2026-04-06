package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andresbott/dashi/internal/sysinfo"
)

func TestSysinfoHandler_GetSysinfo(t *testing.T) {
	h := NewSysinfoHandler(slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/api/v0/widgets/sysinfo", nil)
	rec := httptest.NewRecorder()

	h.GetSysinfo(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var data sysinfo.SystemInfo
	if err := json.NewDecoder(rec.Body).Decode(&data); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(data.Disks) == 0 {
		t.Error("expected at least one disk")
	}
	if data.MemTotal == 0 {
		t.Error("expected non-zero memTotal")
	}
}
