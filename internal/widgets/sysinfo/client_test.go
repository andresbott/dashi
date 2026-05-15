package sysinfo

import (
	"testing"
)

func TestGet(t *testing.T) {
	info, err := Get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(info.Disks) == 0 {
		t.Error("expected at least one disk partition")
	}

	for _, d := range info.Disks {
		if d.Mountpoint == "" {
			t.Error("disk mountpoint should not be empty")
		}
		if d.Total == 0 {
			t.Errorf("disk %s total should not be zero", d.Mountpoint)
		}
		if d.UsedPct < 0 || d.UsedPct > 100 {
			t.Errorf("disk %s usedPct = %f, want 0-100", d.Mountpoint, d.UsedPct)
		}
	}

	if info.MemTotal == 0 {
		t.Error("memTotal should not be zero")
	}
	if info.MemUsedPct < 0 || info.MemUsedPct > 100 {
		t.Errorf("memUsedPct = %f, want 0-100", info.MemUsedPct)
	}

	if info.CPUUsagePct < 0 || info.CPUUsagePct > 100 {
		t.Errorf("cpuUsagePct = %f, want 0-100", info.CPUUsagePct)
	}
	if info.CPUCores == 0 {
		t.Error("cpuCores should not be zero")
	}
	if info.Hostname == "" {
		t.Error("hostname should not be empty")
	}
}
