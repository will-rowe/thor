package colour

import (
	"math"
	"os"
	"testing"
)

var (
	sketch = []uint32{12345, 23456, 34567, 45678, 567895678, 0, math.MaxUint32}
	hex0   = "#00003039"
	sketch2 = []uint32{12345}
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

func TestColourSketchAdjust(t *testing.T) {
	cs := NewColourSketch(sketch2, "coloursketchA")
	if err := cs.Adjust('Q', 100); err == nil {
		t.Fatal("only R/G/B/A should be supported")
	}
	if err := cs.Adjust('R', math.MaxUint8); err == nil {
		t.Fatal("should throw a detailed error when attempting to overflow a RGBA slot")
	}
	if err := cs.Adjust('R', 1); err != nil {
		t.Log(err)
		t.Fatal("should be able to increment R slots by 1")
	}
	_ = cs.Adjust('G', 1)
	_ = cs.Adjust('B', 1)
	_ = cs.Adjust('A', 1)
	rgbLine, err := cs.PrintCSVline(false)
	if err != nil {
		t.Fatal(err)
	}
	if rgbLine != "rgba(58,49,1,1)," {
		t.Fatal("adjust method did not increment each RGBA slot by 1")
	}
	t.Log(rgbLine)
}

func TestPrint(t *testing.T) {
	// check an uninitialised rgb stuct
	emptyColour := &rgba{}
	if err := emptyColour.checker(); err == nil {
		t.Fatal("shouldn't print a hex as there is no rgb values stored")
	}
	cs := NewColourSketch(sketch, "coloursketchA")
	// check the rgb and hex csv line
	hexLine, err := cs.PrintCSVline(true)
	if err != nil {
		t.Fatal(err)
	}
	rgbLine, err := cs.PrintCSVline(false)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(hexLine)
	t.Log(rgbLine)
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
		rgbLine, err := val.PrintCSVline(false)
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
