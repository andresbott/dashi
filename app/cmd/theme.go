package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// canonicalIcons is the list of canonical weather icon names that every theme must provide.
var canonicalIcons = []string{
	"clear-sky",
	"mainly-clear",
	"partly-cloudy",
	"overcast",
	"foggy",
	"drizzle-light",
	"drizzle-moderate",
	"drizzle-dense",
	"freezing-drizzle-light",
	"freezing-drizzle-dense",
	"rain-slight",
	"rain-moderate",
	"rain-heavy",
	"freezing-rain-light",
	"freezing-rain-heavy",
	"snow-slight",
	"snow-moderate",
	"snow-heavy",
	"snow-grains",
	"rain-showers-slight",
	"rain-showers-moderate",
	"rain-showers-violent",
	"snow-showers-slight",
	"snow-showers-heavy",
	"thunderstorm",
	"thunderstorm-hail-slight",
	"thunderstorm-hail-heavy",
	"unknown",
	"sunrise",
	"sunset",
	"wind",
	"humidity",
	"pressure",
	"uv-index",
	"visibility",
	"air-quality",
}

func themeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "theme",
		Short: "manage themes",
	}
	cmd.AddCommand(themeCreateCmd())
	return cmd
}

func themeCreateCmd() *cobra.Command {
	var (
		configFile = "./config.yaml"
		themeType  = "image"
	)

	cmd := &cobra.Command{
		Use:   "create [name]",
		Short: "bootstrap a new theme",
		Long:  "create a new theme directory with a theme.yaml manifest and placeholder files for all weather icons",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if themeType != "image" && themeType != "font" {
				return fmt.Errorf("--type must be \"image\" or \"font\", got %q", themeType)
			}

			cfg, err := getAppCfg(configFile)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			themeDir := filepath.Join(cfg.DataDir, "themes", name)
			if _, err := os.Stat(themeDir); err == nil {
				return fmt.Errorf("theme directory %s already exists", themeDir)
			}

			switch themeType {
			case "image":
				return bootstrapImageTheme(themeDir, name)
			case "font":
				return bootstrapFontTheme(themeDir, name)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&configFile, "config", "c", configFile, "config file path")
	cmd.Flags().StringVarP(&themeType, "type", "t", themeType, "theme type: \"image\" or \"font\"")
	return cmd
}

func bootstrapImageTheme(themeDir, name string) error {
	iconsDir := filepath.Join(themeDir, "widgets", "weather", "icons")
	if err := os.MkdirAll(iconsDir, 0o750); err != nil {
		return fmt.Errorf("creating theme directory: %w", err)
	}

	manifest := fmt.Sprintf(`name: %q
type: icon
description: "Custom weather icons"
icons:
  type: image
`, name)

	if err := os.WriteFile(filepath.Join(themeDir, "theme.yaml"), []byte(manifest), 0o600); err != nil {
		return fmt.Errorf("writing theme.yaml: %w", err)
	}

	placeholder := `<svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 64 64">
  <rect width="64" height="64" fill="#eee" rx="8"/>
  <text x="32" y="36" text-anchor="middle" font-size="10" fill="#999">%s</text>
</svg>
`
	for _, icon := range canonicalIcons {
		content := fmt.Sprintf(placeholder, icon)
		path := filepath.Join(iconsDir, icon+".svg")
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}
	}

	fmt.Printf("Image theme created at %s\n", themeDir)
	fmt.Printf("Replace the SVG files in %s with your custom icons.\n", iconsDir)
	return nil
}

func bootstrapFontTheme(themeDir, name string) error {
	if err := os.MkdirAll(themeDir, 0o750); err != nil {
		return fmt.Errorf("creating theme directory: %w", err)
	}

	manifest := fmt.Sprintf(`name: %q
type: icon
description: "Custom font icon theme"
icons:
  type: font
  # CSS class prefix applied before each icon suffix
  classPrefix: ""
  # Path to the icon font TTF file relative to the theme directory
  fontFile: ""
  # Map canonical weather icon names to font icon class + codepoint
  icons:
    clear-sky:
      class: ""
      codepoint: ""
    mainly-clear:
      class: ""
      codepoint: ""
    partly-cloudy:
      class: ""
      codepoint: ""
    overcast:
      class: ""
      codepoint: ""
    foggy:
      class: ""
      codepoint: ""
    drizzle-light:
      class: ""
      codepoint: ""
    drizzle-moderate:
      class: ""
      codepoint: ""
    drizzle-dense:
      class: ""
      codepoint: ""
    freezing-drizzle-light:
      class: ""
      codepoint: ""
    freezing-drizzle-dense:
      class: ""
      codepoint: ""
    rain-slight:
      class: ""
      codepoint: ""
    rain-moderate:
      class: ""
      codepoint: ""
    rain-heavy:
      class: ""
      codepoint: ""
    freezing-rain-light:
      class: ""
      codepoint: ""
    freezing-rain-heavy:
      class: ""
      codepoint: ""
    snow-slight:
      class: ""
      codepoint: ""
    snow-moderate:
      class: ""
      codepoint: ""
    snow-heavy:
      class: ""
      codepoint: ""
    snow-grains:
      class: ""
      codepoint: ""
    rain-showers-slight:
      class: ""
      codepoint: ""
    rain-showers-moderate:
      class: ""
      codepoint: ""
    rain-showers-violent:
      class: ""
      codepoint: ""
    snow-showers-slight:
      class: ""
      codepoint: ""
    snow-showers-heavy:
      class: ""
      codepoint: ""
    thunderstorm:
      class: ""
      codepoint: ""
    thunderstorm-hail-slight:
      class: ""
      codepoint: ""
    thunderstorm-hail-heavy:
      class: ""
      codepoint: ""
    unknown:
      class: ""
      codepoint: ""
    sunrise:
      class: ""
      codepoint: ""
    sunset:
      class: ""
      codepoint: ""
    wind:
      class: ""
      codepoint: ""
    humidity:
      class: ""
      codepoint: ""
    pressure:
      class: ""
      codepoint: ""
    uv-index:
      class: ""
      codepoint: ""
    visibility:
      class: ""
      codepoint: ""
    air-quality:
      class: ""
      codepoint: ""
`, name)

	if err := os.WriteFile(filepath.Join(themeDir, "theme.yaml"), []byte(manifest), 0o600); err != nil {
		return fmt.Errorf("writing theme.yaml: %w", err)
	}

	fmt.Printf("Font theme created at %s\n", themeDir)
	fmt.Println("Edit theme.yaml to set the font file path, class prefix, and icon mappings.")
	return nil
}
