package pdf

import _ "embed"

//go:embed fonts/verdana.ttf
var verdanaRegular []byte

//go:embed fonts/verdana_bold.ttf
var verdanaBold []byte

//go:embed fonts/verdana_italic.ttf
var verdanaItalic []byte

//go:embed fonts/verdana_bold_italic.ttf
var verdanaBoldItalic []byte
