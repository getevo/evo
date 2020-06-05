package colors

import (
	"errors"
	"strings"
)

var (
	// ErrBadColor is the default bad color error
	ErrBadColor = errors.New("Parsing of color failed, Bad Color")
)

// NamedColor hex value of named colors
type NamedColor struct {
	Hex string
}

var (
	AliceBlue            = NamedColor{"#f0f8ff"}
	AntiqueWhite         = NamedColor{"#faebd7"}
	Aqua                 = NamedColor{"#00ffff"}
	Aquamarine           = NamedColor{"#7fffd4"}
	Azure                = NamedColor{"#f0ffff"}
	Beige                = NamedColor{"#f5f5dc"}
	Bisque               = NamedColor{"#ffe4c4"}
	Black                = NamedColor{"#000000"}
	BlanchedAlmond       = NamedColor{"#ffebcd"}
	Blue                 = NamedColor{"#0000ff"}
	BlueViolet           = NamedColor{"#8a2be2"}
	Brown                = NamedColor{"#a52a2a"}
	BurlyWood            = NamedColor{"#deb887"}
	CadetBlue            = NamedColor{"#5f9ea0"}
	Chartreuse           = NamedColor{"#7fff00"}
	Chocolate            = NamedColor{"#d2691e"}
	Coral                = NamedColor{"#ff7f50"}
	CornflowerBlue       = NamedColor{"#6495ed"}
	Cornsilk             = NamedColor{"#fff8dc"}
	Crimson              = NamedColor{"#dc143c"}
	Cyan                 = NamedColor{"#00ffff"}
	DarkBlue             = NamedColor{"#00008b"}
	DarkCyan             = NamedColor{"#008b8b"}
	DarkGoldenRod        = NamedColor{"#b8860b"}
	DarkGray             = NamedColor{"#a9a9a9"}
	DarkGreen            = NamedColor{"#006400"}
	DarkKhaki            = NamedColor{"#bdb76b"}
	DarkMagenta          = NamedColor{"#8b008b"}
	DarkOliveGreen       = NamedColor{"#556b2f"}
	Darkorange           = NamedColor{"#ff8c00"}
	DarkOrchid           = NamedColor{"#9932cc"}
	DarkRed              = NamedColor{"#8b0000"}
	DarkSalmon           = NamedColor{"#e9967a"}
	DarkSeaGreen         = NamedColor{"#8fbc8f"}
	DarkSlateBlue        = NamedColor{"#483d8b"}
	DarkSlateGray        = NamedColor{"#2f4f4f"}
	DarkTurquoise        = NamedColor{"#00ced1"}
	DarkViolet           = NamedColor{"#9400d3"}
	DeepPink             = NamedColor{"#ff1493"}
	DeepSkyBlue          = NamedColor{"#00bfff"}
	DimGray              = NamedColor{"#696969"}
	DodgerBlue           = NamedColor{"#1e90ff"}
	FireBrick            = NamedColor{"#b22222"}
	FloralWhite          = NamedColor{"#fffaf0"}
	ForestGreen          = NamedColor{"#228b22"}
	Fuchsia              = NamedColor{"#ff00ff"}
	Gainsboro            = NamedColor{"#dcdcdc"}
	GhostWhite           = NamedColor{"#f8f8ff"}
	Gold                 = NamedColor{"#ffd700"}
	GoldenRod            = NamedColor{"#daa520"}
	Gray                 = NamedColor{"#808080"}
	Green                = NamedColor{"#008000"}
	GreenYellow          = NamedColor{"#adff2f"}
	HoneyDew             = NamedColor{"#f0fff0"}
	HotPink              = NamedColor{"#ff69b4"}
	IndianRed            = NamedColor{"#cd5c5c"}
	Indigo               = NamedColor{"#4b0082"}
	Ivory                = NamedColor{"#fffff0"}
	Khaki                = NamedColor{"#f0e68c"}
	Lavender             = NamedColor{"#e6e6fa"}
	LavenderBlush        = NamedColor{"#fff0f5"}
	LawnGreen            = NamedColor{"#7cfc00"}
	LemonChiffon         = NamedColor{"#fffacd"}
	LightBlue            = NamedColor{"#add8e6"}
	LightCoral           = NamedColor{"#f08080"}
	LightCyan            = NamedColor{"#e0ffff"}
	LightGoldenRodYellow = NamedColor{"#fafad2"}
	LightGreen           = NamedColor{"#90ee90"}
	LightGrey            = NamedColor{"#d3d3d3"}
	LightPink            = NamedColor{"#ffb6c1"}
	LightSalmon          = NamedColor{"#ffa07a"}
	LightSeaGreen        = NamedColor{"#20b2aa"}
	LightSkyBlue         = NamedColor{"#87cefa"}
	LightSlateGray       = NamedColor{"#778899"}
	LightSteelBlue       = NamedColor{"#b0c4de"}
	LightYellow          = NamedColor{"#ffffe0"}
	Lime                 = NamedColor{"#00ff00"}
	LimeGreen            = NamedColor{"#32cd32"}
	Linen                = NamedColor{"#faf0e6"}
	Magenta              = NamedColor{"#ff00ff"}
	Maroon               = NamedColor{"#800000"}
	MediumAquaMarine     = NamedColor{"#66cdaa"}
	MediumBlue           = NamedColor{"#0000cd"}
	MediumOrchid         = NamedColor{"#ba55d3"}
	MediumPurple         = NamedColor{"#9370d8"}
	MediumSeaGreen       = NamedColor{"#3cb371"}
	MediumSlateBlue      = NamedColor{"#7b68ee"}
	MediumSpringGreen    = NamedColor{"#00fa9a"}
	MediumTurquoise      = NamedColor{"#48d1cc"}
	MediumVioletRed      = NamedColor{"#c71585"}
	MidnightBlue         = NamedColor{"#191970"}
	MintCream            = NamedColor{"#f5fffa"}
	MistyRose            = NamedColor{"#ffe4e1"}
	Moccasin             = NamedColor{"#ffe4b5"}
	NavajoWhite          = NamedColor{"#ffdead"}
	Navy                 = NamedColor{"#000080"}
	OldLace              = NamedColor{"#fdf5e6"}
	Olive                = NamedColor{"#808000"}
	OliveDrab            = NamedColor{"#6b8e23"}
	Orange               = NamedColor{"#ffa500"}
	OrangeRed            = NamedColor{"#ff4500"}
	Orchid               = NamedColor{"#da70d6"}
	PaleGoldenRod        = NamedColor{"#eee8aa"}
	PaleGreen            = NamedColor{"#98fb98"}
	PaleTurquoise        = NamedColor{"#afeeee"}
	PaleVioletRed        = NamedColor{"#d87093"}
	PapayaWhip           = NamedColor{"#ffefd5"}
	PeachPuff            = NamedColor{"#ffdab9"}
	Peru                 = NamedColor{"#cd853f"}
	Pink                 = NamedColor{"#ffc0cb"}
	Plum                 = NamedColor{"#dda0dd"}
	PowderBlue           = NamedColor{"#b0e0e6"}
	Purple               = NamedColor{"#800080"}
	Red                  = NamedColor{"#ff0000"}
	RosyBrown            = NamedColor{"#bc8f8f"}
	RoyalBlue            = NamedColor{"#4169e1"}
	SaddleBrown          = NamedColor{"#8b4513"}
	Salmon               = NamedColor{"#fa8072"}
	SandyBrown           = NamedColor{"#f4a460"}
	SeaGreen             = NamedColor{"#2e8b57"}
	SeaShell             = NamedColor{"#fff5ee"}
	Sienna               = NamedColor{"#a0522d"}
	Silver               = NamedColor{"#c0c0c0"}
	SkyBlue              = NamedColor{"#87ceeb"}
	SlateBlue            = NamedColor{"#6a5acd"}
	SlateGray            = NamedColor{"#708090"}
	Snow                 = NamedColor{"#fffafa"}
	SpringGreen          = NamedColor{"#00ff7f"}
	SteelBlue            = NamedColor{"#4682b4"}
	Tan                  = NamedColor{"#d2b48c"}
	Teal                 = NamedColor{"#008080"}
	Thistle              = NamedColor{"#d8bfd8"}
	Tomato               = NamedColor{"#ff6347"}
	Turquoise            = NamedColor{"#40e0d0"}
	Violet               = NamedColor{"#ee82ee"}
	Wheat                = NamedColor{"#f5deb3"}
	White                = NamedColor{"#ffffff"}
	WhiteSmoke           = NamedColor{"#f5f5f5"}
	Yellow               = NamedColor{"#ffff00"}
	YellowGreen          = NamedColor{"#9acd32"}
)

// Color is the base color interface from which all others ascribe to
type Color interface {
	ToHEX() *HEXColor
	ToRGB() *RGBColor
	ToRGBA() *RGBAColor
	String() string
	IsLight() bool // http://stackoverflow.com/a/24213274/3158232 and http://www.nbdtech.com/Blog/archive/2008/04/27/Calculating-the-Perceived-Brightness-of-a-Color.aspx
	IsDark() bool  //for perceived luminance, not strict math
}

// Parse parses an unknown color type to it's appropriate type, or returns a ErrBadColor
func Parse(s string) (Color, error) {
	if len(s) < 4 {
		return nil, ErrBadColor
	}

	s = strings.ToLower(s)

	if s[:1] == "#" {
		return ParseHEX(s)
	} else if s[:4] == "rgba" {
		return ParseRGBA(s)
	} else if s[:3] == "rgb" {
		return ParseRGB(s)
	}

	return nil, ErrBadColor
}

// Color return Color instance of named colors
func (c NamedColor) Color() Color {
	color, _ := ParseHEX(c.Hex)
	return color
}

// String return hex value of named colors
func (c NamedColor) String() string {
	return c.Hex
}
