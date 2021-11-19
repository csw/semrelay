package semrelay

import (
	"bytes"
	_ "embed"
	"image"
	"image/png"
)

// This is Semaphore's logo. I take no credit for it and their copyright
// applies.

//go:embed semaphore-sm.png
var Icon []byte
var IconImage *image.NRGBA

func init() {
	img, err := png.Decode(bytes.NewReader(Icon))
	if err != nil {
		panic(err)
	}
	IconImage = img.(*image.NRGBA)
}
