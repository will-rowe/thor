package draw

import (
	"image/color"
	"os"
	"testing"
)

var (
	sketchLength  = 5
	red           = color.RGBA{255, 0, 0, 255}
	green         = color.RGBA{0, 255, 0, 255}
	blue          = color.RGBA{0, 0, 255, 255}
	otu1          = []color.RGBA{red, green, blue, blue, red}
	otu2          = []color.RGBA{blue, green, red, green, green}
	otu3          = []color.RGBA{blue, blue, green, blue, blue}
	longOTUvector = append(otu1, otu2...)
	otus          = [3][]color.RGBA{otu1, otu2, otu3}
)

func TestConstructor(t *testing.T) {
	testImg, err := NewThorPNG(sketchLength, len(otus))
	if err != nil {
		t.Fatal(err)
	}
	if testImg.GetPadding() != (sketchLength - len(otus)) {
		t.Fatal("padding value incorrect")
	}
	if _, err := NewThorPNG(sketchLength, sketchLength+1); err == nil {
		t.Fatal("PNG height can't be greater than PNG width")
	}
}

func TestDrawOTU(t *testing.T) {
	// create the PNG canvas
	testImg, _ := NewThorPNG(sketchLength, len(otus))
	// add the expected OTUS
	for _, otuVector := range otus {
		if err := testImg.DrawOTU(otuVector); err != nil {
			t.Fatal(err)
		}
	}
	// check that unexpected OTUS are handled
	// a shorter vector
	if err := testImg.DrawOTU(otu1[2:]); err == nil {
		t.Fatal("vector length < image width")
	}
	// a longer vector
	if err := testImg.DrawOTU(longOTUvector); err == nil {
		t.Fatal("vector length > image width")
	}
	// too many OTU vectors
	if err := testImg.DrawOTU(otu1); err == nil {
		t.Fatal("image already full")
	}

}

func TestSavePNG(t *testing.T) {
	testImg, _ := NewThorPNG(sketchLength, len(otus))
	for _, otuVector := range otus {
		_ = testImg.DrawOTU(otuVector)
	}
	if err := testImg.Save("./test.png", false); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove("./test.png"); err != nil {
		t.Fatal(err)
	}
}

// test padding, printing, etc.
