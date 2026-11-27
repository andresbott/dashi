package themes

// EmberID is the ID of the builtin "Ember" theme: warm paper in light mode,
// warm charcoal with amber accents in dark mode.
const EmberID = "ember"

// builtins maps every builtin theme ID to its constructor. Builtins live in
// code, are never written to disk, and refuse every write with ErrBuiltin;
// Copy is how they get customised. Their IDs must not be 6 characters long,
// so they can never collide with a randomID.
var builtins = map[string]func() Theme{
	DefaultID: DefaultTheme,
	EmberID:   EmberTheme,
}

// IsBuiltin reports whether id names a builtin theme.
func IsBuiltin(id string) bool {
	_, ok := builtins[id]
	return ok
}

// Builtin returns the builtin theme with the given ID.
func Builtin(id string) (Theme, bool) {
	fn, ok := builtins[id]
	if !ok {
		return Theme{}, false
	}
	return fn(), true
}

// EmberTheme returns the builtin Ember theme over a plain background.
func EmberTheme() Theme {
	return Theme{
		ID:       EmberID,
		Name:     "Ember",
		DarkMode: true,
		Colors: Colors{
			Light: Palette{
				BG: "#f4ecdf", Surface: "#fbf6ee", Border: "#e2d5c0",
				Text: "#3b2f25", TextMuted: "#6e5e4f",
				Primary: "#b5501a", OnPrimary: "#ffffff",
				Secondary: "#2f7d78", Success: "#5b8a3a", Warning: "#b7791f", Danger: "#b8322a",
			},
			Dark: Palette{
				BG: "#1c1714", Surface: "#26201b", Border: "#3a3029",
				Text: "#ede3d6", TextMuted: "#a8998a",
				Primary: "#f0a030", OnPrimary: "#1c1714",
				Secondary: "#5fb0a6", Success: "#8fb85a", Warning: "#e8b64a", Danger: "#e5534b",
			},
		},
	}
}
