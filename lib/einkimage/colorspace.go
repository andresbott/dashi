package einkimage

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// hexToRGB parses a hex color string into RGB components.
// Accepts "#RGB", "RGB", "#RRGGBB", and "RRGGBB" forms.
func hexToRGB(hex string) ([3]uint8, error) {
	s := strings.TrimPrefix(hex, "#")

	if len(s) == 3 {
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	}
	if len(s) != 6 {
		return [3]uint8{}, fmt.Errorf("invalid hex color: %q", hex)
	}

	var out [3]uint8
	for i := 0; i < 3; i++ {
		v, err := strconv.ParseUint(s[i*2:i*2+2], 16, 8)
		if err != nil {
			return [3]uint8{}, fmt.Errorf("invalid hex color: %q", hex)
		}
		out[i] = uint8(v)
	}
	return out, nil
}

// clampByte clamps a float to [0, 255] and rounds to the nearest integer,
// matching the JS reference's Math.round + clamp behavior. NaN and Inf
// produce 0.
func clampByte(v float64) uint8 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(math.Round(v))
}

// luma709 computes Rec.709 luma from 8-bit RGB components.
func luma709(r, g, b uint8) float64 {
	return 0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)
}

// RGBToLab converts 8-bit sRGB to CIE L*a*b* (D65). Returns [L, a, b].
func RGBToLab(r, g, b uint8) [3]float64 {
	x, y, z := rgbToXYZ(r, g, b)
	return xyzToLab(x, y, z)
}

// LabToRGB converts a CIE L*a*b* triple back to 8-bit sRGB.
func LabToRGB(lab [3]float64) (r, g, b uint8) {
	x, y, z := labToXYZ(lab[0], lab[1], lab[2])
	return xyzToRGB(x, y, z)
}

// DeltaE computes the Euclidean distance between two LAB colors (CIE76).
func DeltaE(lab1, lab2 [3]float64) float64 {
	dl := lab1[0] - lab2[0]
	da := lab1[1] - lab2[1]
	db := lab1[2] - lab2[2]
	return math.Sqrt(dl*dl + da*da + db*db)
}

func srgbToLinear(v float64) float64 {
	if v > 0.04045 {
		return math.Pow((v+0.055)/1.055, 2.4)
	}
	return v / 12.92
}

func linearToSRGB(v float64) float64 {
	if v > 0.0031308 {
		return 1.055*math.Pow(v, 1.0/2.4) - 0.055
	}
	return 12.92 * v
}

func rgbToXYZ(r, g, b uint8) (float64, float64, float64) {
	rn := srgbToLinear(float64(r) / 255)
	gn := srgbToLinear(float64(g) / 255)
	bn := srgbToLinear(float64(b) / 255)
	return (rn*0.4124564 + gn*0.3575761 + bn*0.1804375) * 100,
		(rn*0.2126729 + gn*0.7151522 + bn*0.072175) * 100,
		(rn*0.0193339 + gn*0.119192 + bn*0.9503041) * 100
}

func xyzToLab(x, y, z float64) [3]float64 {
	f := func(t float64) float64 {
		if t > 0.008856 {
			return math.Pow(t, 1.0/3.0)
		}
		return 7.787*t + 16.0/116.0
	}
	xn := f(x / 95.047)
	yn := f(y / 100)
	zn := f(z / 108.883)
	return [3]float64{116*yn - 16, 500 * (xn - yn), 200 * (yn - zn)}
}

func labToXYZ(l, a, b float64) (float64, float64, float64) {
	y := (l + 16) / 116
	x := a/500 + y
	z := y - b/200
	f := func(t float64) float64 {
		if t > 0.206897 {
			return t * t * t
		}
		return (t - 16.0/116.0) / 7.787
	}
	return f(x) * 95.047, f(y) * 100, f(z) * 108.883
}

func xyzToRGB(x, y, z float64) (uint8, uint8, uint8) {
	xn := x / 100
	yn := y / 100
	zn := z / 100
	r := linearToSRGB(xn*3.2404542 + yn*-1.5371385 + zn*-0.4985314)
	g := linearToSRGB(xn*-0.969266 + yn*1.8760108 + zn*0.041556)
	b := linearToSRGB(xn*0.0556434 + yn*-0.2040259 + zn*1.0572252)
	return clampByte(r * 255), clampByte(g * 255), clampByte(b * 255)
}
