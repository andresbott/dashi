package einkimage

import "sort"

// Reference 8x8 Bayer matrix matching src/dither/functions/bayer-matrix.ts.
var bayer8x8 = [8][8]int{
	{0, 48, 12, 60, 3, 51, 15, 63},
	{32, 16, 44, 28, 35, 19, 47, 31},
	{8, 56, 4, 52, 11, 59, 7, 55},
	{40, 24, 36, 20, 43, 27, 39, 32},
	{2, 50, 14, 62, 1, 49, 13, 61},
	{34, 18, 46, 30, 33, 17, 45, 29},
	{10, 58, 6, 54, 9, 57, 5, 53},
	{42, 26, 38, 22, 41, 25, 37, 21},
}

// bayerMatrix returns a 2D int matrix of the requested size, derived from
// bayer8x8. For sub-8 sizes, the JS code samples bigMatrix[x][y] and then
// re-ranks values into the dense range [0, w*h).
func bayerMatrix(width, height int) [][]int {
	if width > 8 {
		width = 8
	}
	if height > 8 {
		height = 8
	}
	if width == 8 && height == 8 {
		out := make([][]int, 8)
		for y := range out {
			row := make([]int, 8)
			for x := 0; x < 8; x++ {
				row[x] = bayer8x8[y][x]
			}
			out[y] = row
		}
		return out
	}

	// JS samples as matrix[y][x] = bigMatrix[x][y] (note index swap).
	m := make([][]int, height)
	for y := 0; y < height; y++ {
		row := make([]int, width)
		for x := 0; x < width; x++ {
			row[x] = bayer8x8[x][y]
		}
		m[y] = row
	}

	// Re-rank into dense [0, w*h).
	flat := make([]int, 0, width*height)
	for _, row := range m {
		flat = append(flat, row...)
	}
	sort.Ints(flat)
	rank := make(map[int]int, len(flat))
	for i, v := range flat {
		if _, ok := rank[v]; !ok {
			rank[v] = i
		}
	}
	for y, row := range m {
		for x, v := range row {
			m[y][x] = rank[v]
		}
	}
	return m
}

// orderedDither adds a Bayer threshold per pixel then snaps to the nearest
// palette color.
func orderedDither(buf []uint8, width, height int, palette Palette, size [2]int, matching ColorMatchingMode) {
	w, h := size[0], size[1]
	if w == 0 {
		w = 4
	}
	if h == 0 {
		h = 4
	}
	matrix := bayerMatrix(w, h)
	div := float64(len(matrix) * len(matrix[0]))
	const threshold = 256.0

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			mx := x % len(matrix[0])
			my := y % len(matrix)
			factor := (float64(matrix[my][mx]) / div) - 0.5

			idx := (y*width + x) * 4
			r := clampByte(float64(buf[idx]) + factor*threshold)
			g := clampByte(float64(buf[idx+1]) + factor*threshold)
			b := clampByte(float64(buf[idx+2]) + factor*threshold)
			newPixel := findClosestPaletteColor([3]uint8{r, g, b}, palette, matching)
			buf[idx] = newPixel[0]
			buf[idx+1] = newPixel[1]
			buf[idx+2] = newPixel[2]
		}
	}
}
