package colour

import (
	"math"
	"os"
	"testing"
)

var (
	sketch = []uint32{12345, 23456, 34567, 45678, 567895678, 0, math.MaxUint32}
	hex0   = "#00003039"
)

func TestColourSketch(t *testing.T) {
	cs := NewColourSketch(sketch, "coloursketchA")
	for i, colour := range cs.Colours {
		t.Log(i, colour)
	}
	// the 6th colour in the sketch should be set to 0s (black)
	if cs.Colours[5].printRGBA() != "rgba(0,0,0,0)" {
		t.Fatal("failed to colorsketch")
	}
	// the 7th colour in the sketch sketch should be set to 255s (white)
	if cs.Colours[6].printRGBA() != "rgba(255,255,255,255)" {
		t.Fatal("failed to colorsketch")
	}
}

func TestPrint(t *testing.T) {
	// check an uninitialised rgb stuct
	emptyColour := &rgba{}
	if err := emptyColour.checker(); err == nil {
		t.Fatal("shouldn't print a hex as there is no rgb values stored")
	}
	cs := NewColourSketch(sketch, "coloursketchA")
	// check the rgb and hex csv line
	hexLine, err := cs.Print(true)
	if err != nil {
		t.Fatal(err)
	}
	rgbLine, err := cs.Print(false)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(hexLine)
	t.Log(rgbLine)
}

func TestRGBA2Hex(t *testing.T) {
	// check that the hex encoding works
	cs := NewColourSketch(sketch, "coloursketchA")
	if cs.Colours[0].Hex != hex0 {
		t.Fatal("hex encoding failed")
	}
	converted, err := Hex2rgba(cs.Colours[0].Hex)
	if err != nil {
		t.Fatal(err)
	}
	if converted.printRGBA() != cs.Colours[0].printRGBA() {
		t.Fatal("rgba converted hex does not match original rgba value")
	}
}

// test the lshEnsemble dump and load methods
func Test_ColourSketchStoreDump(t *testing.T) {
	// create the store
	css := make(ColourSketchStore)
	// add a coloursketch
	cs := NewColourSketch(sketch, "coloursketchA")
	css[cs.Id] = cs
	// dump
	if err := css.Dump("./css.thor"); err != nil {
		t.Fatal(err)
	}
	// load
	css2 := make(ColourSketchStore)
	if err := css2.Load("./css.thor"); err != nil {
		t.Fatal(err)
	}
	// check the dump/load worked
	for key, val := range css2 {
		if key != val.Id {
			t.Fatal("id mismatch")
		}
		rgbLine, err := val.Print(false)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(rgbLine)
	}
	if css2.GetSketchLength() != 7 {
		t.Fatal("cannot access sketch length from coloursketch store")
	}
	// rm file
	if err := os.Remove("./css.thor"); err != nil {
		t.Fatal(err)
	}
}
