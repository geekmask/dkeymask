package theme

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"

	"image/color"
)

type CustomTheme struct{}

var _ fyne.Theme = (*CustomTheme)(nil)

func (*CustomTheme) Font(s fyne.TextStyle) fyne.Resource {
	return resourcePingfangTtf
}
func (*CustomTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	// if n == theme.ColorNameInputBackground {
	// 	return color.White
	// }
	v = theme.VariantDark
	return theme.DefaultTheme().Color(n, v)
}

func (*CustomTheme) Size(n fyne.ThemeSizeName) float32 {
	if n == theme.SizeNameInputBorder {
		return 1
	}
	return theme.DefaultTheme().Size(n)
}

func (*CustomTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(n)
}

func WindowIcon() fyne.Resource {
	return resourceLogoPng
}
