package colour

import (
	"math"
	"testing"
)

var (
	histosketch = []uint32{12345, 23456, 34567, 45678, 567895678, 0, math.MaxUint32}
	hex0 = "#003039"
)

func TestColourHistosketch(t *testing.T) {
	rgb := ColourHistosketch(histosketch)
	for i, colour := range rgb {
		t.Log(i, colour)
	}
	// the 6th colour sketch should be set to 0s (black)
	if rgb[5].printRGBA() != "rgba(0,0,0,0)" {
		t.Fatal("failed to colorsketch")
	}
	// the 7th colour sketch should be set to 255s (white)
	if rgb[6].printRGBA() != "rgba(255,255,255,255)" {
		t.Fatal("failed to colorsketch")
	}
}

func TestPrint(t *testing.T) {
	// check an uninitialised rgb stuct
	emptyColour := &rgba{}
	if err := emptyColour.checker(); err == nil {
		t.Fatal("shouldn't print a hex as there is no rgb values stored")
	}
	rgb := ColourHistosketch(histosketch)
	// check the rgb and hex csv line
	hexLine, err := rgb.Print(true)
	if err != nil {
		t.Fatal(err)
	}
	rgbLine, err := rgb.Print(false)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(hexLine)
	t.Log(rgbLine)
}
