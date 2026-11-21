package dashboard

import (
	"errors"
	"fmt"
	"slices"

	"github.com/andresbott/dashi/lib/einkimage"
)

// ErrInvalidDither is returned by Create, Update and ImportZip when a
// dashboard's Dither is not one of DitherModes.
var ErrInvalidDither = errors.New("invalid dither mode")

// DitherAuto is the Dither value that lets the renderer pick the processing
// for each image (einkimage.SuggestProcessingOptions).
const DitherAuto = "auto"

// DitherModes lists the non-empty values Dashboard.Dither accepts: "auto"
// followed by the einkimage processing presets. Empty (plain dithering) is
// always accepted as well.
func DitherModes() []string {
	return append([]string{DitherAuto}, einkimage.GetProcessingPresetNames()...)
}

func validateDither(mode string) error {
	if mode == "" || slices.Contains(DitherModes(), mode) {
		return nil
	}
	return fmt.Errorf("%w %q: want empty or one of %v", ErrInvalidDither, mode, DitherModes())
}
