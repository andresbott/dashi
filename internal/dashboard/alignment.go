package dashboard

// ContainerPlacement is the CSS that positions the dashboard container (the
// wrapper around all rows) inside the viewport. Both server renders apply it
// the same way, so a dashboard sits in the same place in the browser viewer
// and in the e-ink PNG:
//
//   - MarginLeft and MarginRight go on the container itself. It is full width
//     up to its max-width, so the auto side absorbs the leftover space and
//     moves the whole container; the rows inside keep filling it.
//   - JustifyContent goes on a full-viewport-height flex column around the
//     container, moving it to the top, middle or bottom.
type ContainerPlacement struct {
	MarginLeft     string
	MarginRight    string
	JustifyContent string
}

// PlaceContainer maps a dashboard's stored alignment (Container.HorizontalAlign
// left|center|right, Container.VerticalAlign top|center|bottom) to CSS. An
// empty or unknown value falls back to centred horizontally and top vertically,
// the defaults both renders have always used.
func PlaceContainer(horizontal, vertical string) ContainerPlacement {
	p := ContainerPlacement{MarginLeft: "auto", MarginRight: "auto", JustifyContent: "flex-start"}
	switch horizontal {
	case "left":
		p.MarginLeft = "0"
	case "right":
		p.MarginRight = "0"
	}
	switch vertical {
	case "center":
		p.JustifyContent = "center"
	case "bottom":
		p.JustifyContent = "flex-end"
	}
	return p
}
