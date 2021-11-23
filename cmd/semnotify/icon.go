package main

import "image"

type DBusIcon struct {
	Width         int
	Height        int
	Rowstride     int
	HasAlpha      bool
	BitsPerSample int
	Channels      int
	Data          []uint8
}

func buildIcon(img *image.NRGBA) *DBusIcon {
	return &DBusIcon{
		Width:         img.Rect.Dx(),
		Height:        img.Rect.Dy(),
		Rowstride:     4 * img.Rect.Dx(),
		HasAlpha:      true,
		BitsPerSample: 8,
		Channels:      4,
		Data:          img.Pix,
	}
}
