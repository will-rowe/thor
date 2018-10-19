package colour

import (
	"testing"
)

var (
	histosketch = []uint64{12345, 23456, 34567, 45678, 56789}
	hex0 = "#003039"
)

func TestColourHistosketch(t *testing.T) {
	rgb := ColourHistosketch(histosketch)
	for i, colour := range rgb {
		t.Log(i, colour)
	}
}

func TestPrints(t *testing.T) {
	// check an uninitialised rgb stuct
	emptyColour := &rgb{}
	if err := emptyColour.checker(); err == nil {
		t.Fatal("shouldn't print a hex as there is no rgb values stored")
	}
	rgb := ColourHistosketch(histosketch)
	// check the individual prints
	hexString := rgb[0].printHex()
	rgbString := rgb[0].printRGB()
	t.Log(hexString, rgbString)
	// check the rgb and hex csv line
	hexLine, err := rgb.Print(true)
	if err != nil {
		t.Fatal(err)
	}
	rgbLine, err := rgb.Print(false)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(hexLine, rgbLine)
	// TODO: check the colours actual match and make sense!


	


}
