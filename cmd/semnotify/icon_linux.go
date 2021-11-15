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
	raw := []uint8{}
	for i := 0; i < img.Rect.Dx()*img.Rect.Dy(); i++ {
		raw = append(raw, img.Pix[i*4:i*4+3]...)
	}
	return &DBusIcon{
		Width:         img.Rect.Dx(),
		Height:        img.Rect.Dy(),
		Rowstride:     3 * img.Rect.Dx(),
		HasAlpha:      false,
		BitsPerSample: 8,
		Channels:      3,
		Data:          raw,
	}
}
