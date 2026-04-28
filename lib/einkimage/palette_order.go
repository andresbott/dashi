package einkimage

import "sort"

// canonicalColorOrder mirrors the JS CANONICAL_COLOR_ORDER constant.
var canonicalColorOrder = []string{
	"black", "white", "blue", "green", "red", "orange", "yellow",
	"gameboy0", "gameboy1", "gameboy2", "gameboy3",
}

func roleRank(role string, fallbackIndex int) int {
	for i, name := range canonicalColorOrder {
		if name == role {
			return i
		}
	}
	return len(canonicalColorOrder) + fallbackIndex
}

// sortByCanonicalOrder returns a copy of entries sorted by role.
// Unknown roles preserve their relative input order via a stable sort.
func sortByCanonicalOrder(entries Palette) Palette {
	type indexed struct {
		entry PaletteEntry
		index int
	}
	pairs := make([]indexed, len(entries))
	for i, e := range entries {
		pairs[i] = indexed{entry: e, index: i}
	}
	sort.SliceStable(pairs, func(i, j int) bool {
		return roleRank(pairs[i].entry.Name, pairs[i].index) <
			roleRank(pairs[j].entry.Name, pairs[j].index)
	})
	out := make(Palette, len(entries))
	for i, p := range pairs {
		out[i] = p.entry
	}
	return out
}

// AlignDeviceColors returns src with each entry's DeviceColor replaced by the
// DeviceColor of the entry in target that has the same Name (role). Entries
// with no matching role in target keep their existing DeviceColor.
func AlignDeviceColors(src, target Palette) Palette {
	deviceByRole := make(map[string][3]uint8, len(target))
	for _, e := range target {
		deviceByRole[e.Name] = e.DeviceColor
	}
	out := make(Palette, len(src))
	for i, e := range src {
		out[i] = e
		if dc, ok := deviceByRole[e.Name]; ok {
			out[i].DeviceColor = dc
		}
	}
	return out
}
