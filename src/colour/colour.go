// colour contains the types/methods/functions to encode sketches as rgba values

package colour

import (
	"fmt"
	"image/color"
	"io/ioutil"
	"math"

	"gopkg.in/vmihailenco/msgpack.v2"
)

// colourSketchStore is a struct to hold and query a set of coloured sketches
type ColourSketchStore map[string]*colourSketch

// Dump a colourSketchStore to disk
func (ColourSketchStore *ColourSketchStore) Dump(path string) error {
	b, err := msgpack.Marshal(ColourSketchStore)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, b, 0644)
}

// Load a colourSketchStore from disk
func (ColourSketchStore *ColourSketchStore) Load(path string) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = msgpack.Unmarshal(b, ColourSketchStore)
	if err != nil {
		return err
	}
	return nil
}

// GetSketchLength returns the number of elements per sketch
func (ColourSketchStore *ColourSketchStore) GetSketchLength() int {
	var key string
	for key = range *ColourSketchStore {
		break
	}
	return len((*ColourSketchStore)[key].Colours)
}

// colourSketch is a struct to hold a colour encoded sketch
type colourSketch struct {
	Colours []rgba
	Id      string
}

// CopySketch returns a copy of the colourSketch
func (cs *colourSketch) CopySketch() *colourSketch {
	c := make([]rgba, len(cs.Colours))
	for i := 0; i < len(c); i++ {
		c[i] = cs.Colours[i]
	}
	return &colourSketch{
		Colours: c,
		Id:      cs.Id,
	}
}

// PrintCSVline is a method to print the coloured sketch as a csv line (either in rgb or hex)
func (colourSketch *colourSketch) PrintCSVline(printHex bool) (string, error) {
	if colourSketch.Id == "" {
		return "", fmt.Errorf("no ID is set for this colour sketch")
	}
	var line string
	for _, value := range colourSketch.Colours {
		// make sure the rgb is okay to use
		if err := value.checker(); err != nil {
			return "", err
		}
		if printHex {
			line += value.Hex + ","
		} else {
			line += value.printRGBA() + ","
		}
	}
	return line, nil
}

// PrintPNGline is a method to print the coloured sketch as a line for PNG conversion
func (colourSketch *colourSketch) PrintPNGline() ([]color.RGBA, error) {
	if colourSketch.Id == "" {
		return nil, fmt.Errorf("no ID is set for this colour sketch")
	}
	line := make([]color.RGBA, len(colourSketch.Colours))
	for i, colour := range colourSketch.Colours {
		// make sure the rgb is okay to use
		if err := colour.checker(); err != nil {
			return nil, err
		}
		line[i] = colour.RGBA
	}
	return line, nil
}

// Adjust is a method to increment a RGBA slot in each element of a colourSketch
func (colourSketch *colourSketch) Adjust(slot rune, increment uint8) error {
	var overflowCheck uint16
	switch slot {
	case 'R':
		for i := range colourSketch.Colours {
			overflowCheck = uint16(colourSketch.Colours[i].RGBA.R) + uint16(increment)
			if overflowCheck > math.MaxUint8 {
				return fmt.Errorf("overflow error: can't increment curent value (%d) by %d", colourSketch.Colours[i].RGBA.R, increment)
			}
			colourSketch.Colours[i].RGBA.R += increment
		}
	case 'G':
		for i := range colourSketch.Colours {
			overflowCheck = uint16(colourSketch.Colours[i].RGBA.G) + uint16(increment)
			if overflowCheck > math.MaxUint8 {
				return fmt.Errorf("overflow error: can't increment curent value (%d) by %d", colourSketch.Colours[i].RGBA.G, increment)
			}
			colourSketch.Colours[i].RGBA.G += increment
		}
	case 'B':
		for i := range colourSketch.Colours {
			overflowCheck = uint16(colourSketch.Colours[i].RGBA.B) + uint16(increment)
			if overflowCheck > math.MaxUint8 {
				return fmt.Errorf("overflow error: can't increment curent value (%d) by %d", colourSketch.Colours[i].RGBA.B, increment)
			}
			colourSketch.Colours[i].RGBA.B += increment
		}
	case 'A':
		for i := range colourSketch.Colours {
			overflowCheck = uint16(colourSketch.Colours[i].RGBA.A) + uint16(increment)
			if overflowCheck > math.MaxUint8 {
				return fmt.Errorf("overflow error: can't increment curent value (%d) by %d", colourSketch.Colours[i].RGBA.A, increment)
			}
			colourSketch.Colours[i].RGBA.A += increment
		}
	default:
		return fmt.Errorf("unknown slot (%v): only R/G/B/A supported", slot)
	}
	return nil
}

// parcel helps to parcel colour sketches and error messages for sending over a channel
type parcel struct {
	cs  *colourSketch
	err error
}

// Unpack is a method to give the colourSketch and error that a parcel contains
func (parcel *parcel) Unpack() (*colourSketch, error) {
	return parcel.cs, parcel.err
}

// colourSketchChan sends parcels
type colourSketchChan chan parcel

// Send is a method to send a coloursketch and an error via the colourSketchChan
func (colourSketchChan colourSketchChan) Send(x *colourSketch, y error) {
	colourSketchChan <- parcel{
		cs:  x,
		err: y,
	}
}

// rgb is a struct to hold the colour information for one sketch element
type rgba struct {
	RGBA color.RGBA
	Hex  string
}

// checker is a method to check the rgb can be used
func (rgba *rgba) checker() error {
	if rgba.Hex == "" {
		return fmt.Errorf("this is an uninitialised rgba")
	}
	return nil
}

// printRGB is a method to convert an rbga struct to a rgba string
func (rgba *rgba) printRGBA() string {
	return fmt.Sprintf("rgba(%d,%d,%d,%d)", rgba.RGBA.R, rgba.RGBA.G, rgba.RGBA.B, rgba.RGBA.A)
}

// printHex is a method to convert an rgba struct to a hex string
func (rgba *rgba) printHex() string {
	return fmt.Sprintf("#%02X%02X%02X%02X", rgba.RGBA.R, rgba.RGBA.G, rgba.RGBA.B, rgba.RGBA.A)
}

// NewColourSketch is is the colourSketch constructor function
func NewColourSketch(sketch []uint32, v string) *colourSketch {
	// prepare the coloured sketch
	c := make([]rgba, len(sketch))
	// iterate over the values in the sketch and convert them to rgba
	for i := 0; i < len(sketch); i++ {
		c[i] = getRGBA(sketch[i])
	}
	return &colourSketch{
		Colours: c,
		Id:      v,
	}
}

// NewColourSketchChan is a constructor function to create a channel for sending colour sketches
func NewColourSketchChan() colourSketchChan {
	return make(colourSketchChan)
}

// getRGBA is a helper function to convert a uint32 to an RGBA colour
// begin loading the RGBA from the most significant bit
func getRGBA(element uint32) rgba {
	colour := color.RGBA{
		R: uint8(0xFF & element),
		G: uint8(0xFF & (element >> 8)),
		B: uint8(0xFF & (element >> 16)),
		A: uint8(0xFF & (element >> 24)),
	}
	rgba := rgba{
		RGBA: colour,
	}
	rgba.Hex = rgba.printHex()
	return rgba
}
