package einkimage

import "testing"

func TestSortByCanonicalOrder(t *testing.T) {
	// Given entries in arbitrary order, result must be in canonical order.
	in := Palette{
		{Name: "yellow"},
		{Name: "black"},
		{Name: "red"},
		{Name: "custom"}, // unknown role preserves original relative index
		{Name: "white"},
	}
	got := sortByCanonicalOrder(in)
	wantNames := []string{"black", "white", "red", "yellow", "custom"}
	for i, w := range wantNames {
		if got[i].Name != w {
			t.Errorf("pos %d: got %q, want %q (full=%v)",
				i, got[i].Name, w, namesOf(got))
		}
	}
}

func TestSortByCanonicalOrderPreservesOriginalIndexForUnknowns(t *testing.T) {
	// Two unknown roles must keep their input order relative to each other.
	in := Palette{
		{Name: "foo"},
		{Name: "bar"},
	}
	got := sortByCanonicalOrder(in)
	if got[0].Name != "foo" || got[1].Name != "bar" {
		t.Errorf("unknown role ordering broken: %v", namesOf(got))
	}
}

func TestAlignDeviceColors(t *testing.T) {
	src := Palette{
		{Name: "red", Color: [3]uint8{100, 0, 0}, DeviceColor: [3]uint8{250, 0, 0}},
		{Name: "green", Color: [3]uint8{0, 100, 0}, DeviceColor: [3]uint8{0, 250, 0}},
	}
	target := Palette{
		{Name: "red", Color: [3]uint8{255, 0, 0}, DeviceColor: [3]uint8{255, 0, 0}},
		{Name: "green", Color: [3]uint8{0, 255, 0}, DeviceColor: [3]uint8{0, 255, 0}},
	}
	aligned := AlignDeviceColors(src, target)
	if aligned[0].DeviceColor != ([3]uint8{255, 0, 0}) {
		t.Errorf("red device color not aligned: %v", aligned[0].DeviceColor)
	}
	if aligned[0].Color != ([3]uint8{100, 0, 0}) {
		t.Errorf("red calibrated color must stay on src: %v", aligned[0].Color)
	}
}

func namesOf(p Palette) []string {
	out := make([]string, len(p))
	for i, e := range p {
		out[i] = e.Name
	}
	return out
}
