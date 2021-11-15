package semrelay

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
)

var Icon []byte
var IconImage *image.NRGBA

// This is Semaphore's logo. I take no credit for it and their copyright
// applies.

func init() {
	var err error
	Icon, err = base64.StdEncoding.DecodeString(`
iVBORw0KGgoAAAANSUhEUgAAABQAAAAUCAYAAACNiR0NAAABhGlDQ1BJQ0MgcHJvZmlsZQAAKJF9
kT1Iw0AcxV9TS0WrHewg4pChOlkQFXHUKhShQqgVWnUwufQLmjQkKS6OgmvBwY/FqoOLs64OroIg
+AHi5uak6CIl/i8ttIjx4Lgf7+497t4BQr3MNKtrHNB020wl4mImuyoGX9GLMPoxgIDMLGNOkpLw
HF/38PH1LsazvM/9OfrUnMUAn0g8ywzTJt4gnt60Dc77xBFWlFXic+Ixky5I/Mh1pclvnAsuCzwz
YqZT88QRYrHQwUoHs6KpEU8RR1VNp3wh02SV8xZnrVxlrXvyF4Zy+soy12kOI4FFLEGCCAVVlFCG
jRitOikWUrQf9/APuX6JXAq5SmDkWEAFGmTXD/4Hv7u18pMTzaRQHAi8OM7HCBDcBRo1x/k+dpzG
CeB/Bq70tr9SB2Y+Sa+1tegREN4GLq7bmrIHXO4Ag0+GbMqu5Kcp5PPA+xl9UxYYuAV61pq9tfZx
+gCkqavkDXBwCIwWKHvd493dnb39e6bV3w8M9nJ+oCssEgAAAAZiS0dEAP8A/wD/oL2nkwAAAAlw
SFlzAAAuIwAALiMBeKU/dgAAAAd0SU1FB+ULDw8eKE9G0c4AAAAZdEVYdENvbW1lbnQAQ3JlYXRl
ZCB3aXRoIEdJTVBXgQ4XAAACz0lEQVQ4y62US2xMURjHf985t63O1KCotl6JRMQjiJCSYlE2EgtE
tLvGQpQQG68FtiIkCCLVNJEGiQ0LIgSJeg2KEm/1ClE6JqataTt1H8didGZqZgRxdve75/7O9/9/
93/EGGP4j0tle+F1duDeu4Xp7koWHRvn3WvwvOxEk2HZd26Y7hWjTM+KESZ2vDZR7z19wvQuLzWx
LdXGefsq06dGMkl2797E3rUMAFEK1zVIziCUG0E8iUvbehpr5pw/kOw4uE+bkwo8DyUGcdoxJg5z
h05BT5icUbEVN8zFvn8bGVSI23gOLuwGJJNBeMOnk7ejHsfn4/KzIMFPL6icOJ9JJeOSQOdRM86e
KsR4YExWmCqejbX9MHbhMHZfP8Gh0BMMEHoU5XDJuiTQbbqKcR36UAmc9qNX7kFPmobX2Y4uHkks
EGDntWPUfXke9xh4H+ugq7cHf15+HGgtrsJ8DeE1NZDan15fS055RT+z9187zpHwEwQFCIsKSthX
sQp/Xn5yny4uRS+tThOpp8xIq5X6CxGJWwAw1jeEwAB/+pQlJzfdte5oWq2tu72fx6Heb9iu0x9o
wiHsk3UoY5InTK9EBQb3gz382MLe1gfgSaKXUx0fqLlwkEhXZ8pQglfQTUcxIniAmlVN7vrtRDyb
nRdrGTewiNauCPXhlyAgPzvsS8Sl6GdytJUE6rJ5mAYBEay5NVirNxO2e9jQWE9j9At8fZOYKCm5
WlBQxNNYO2vHlFEwwBff0xc9t+UZku9DlY6mLRphzZU6grEIKstdtK5kKhvLK7GUxmBQolKSAujx
ExObD907QzAWicuSxH+dGEWFv4hNc6uwlP7ZufwSvdQouw4N4ReJ521jZjPcF+B1RxsHWh8C8Lgn
ghb1myynFrTF1YUbOPsyiOt51JQtQYnw3bVpOR8iV2mqxpcjIhmB8jc3tmcMKgvon4B/sn4A7EVS
OcW9vdYAAAAASUVORK5CYII=`)
	if err != nil {
		panic(err)
	}
	img, err := png.Decode(bytes.NewReader(Icon))
	if err != nil {
		panic(err)
	}
	IconImage = img.(*image.NRGBA)
}
