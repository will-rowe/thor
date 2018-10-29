// colour contains the types/methods/functions to encode sketches as rgba values

package colour

import (
	"encoding/hex"
	"errors"
	"fmt"
	"gopkg.in/vmihailenco/msgpack.v2"
	"io/ioutil"
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

// Print is a method to print the coloured sketch as a csv line (either in rgb or hex)
func (colourSketch *colourSketch) Print(printHex bool) (string, error) {
	if colourSketch.Id == "" {
		return "", errors.New("no ID is set for this colour sketch")
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
	R   uint8
	G   uint8
	B   uint8
	A   uint8
	Hex string
}

// checker is a method to check the rgb can be used
func (rgba *rgba) checker() error {
	if rgba.Hex == "" {
		return errors.New("this is an uninitialised rgba")
	}
	return nil
}

// printRGB is a method to convert an rbga struct to a rgba string
func (rgba *rgba) printRGBA() string {
	return fmt.Sprintf("rgba(%d,%d,%d,%d)", rgba.R, rgba.G, rgba.B, rgba.A)
}

// printHex is a method to convert an rgba struct to a hex string
func (rgba *rgba) printHex() string {
	return fmt.Sprintf("#%02X%02X%02X%02X", rgba.R, rgba.G, rgba.B, rgba.A)
}

// NewColourSketch is is the colourSketch constructor
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

// Hex2rgba is a function to convert a hex string back to an rgba struct
func Hex2rgba(h string) (rgba, error) {
	trimmedH := h
	if trimmedH[0] == '#' {
		trimmedH = h[1:]
	}
	if len(trimmedH) != 8 {
		return rgba{}, errors.New(fmt.Sprintf("Invalid hex string: %s", h))
	}
	decodedR, err := hex.DecodeString(trimmedH[0:2])
	if err != nil {
		return rgba{}, err
	}
	decodedG, err := hex.DecodeString(trimmedH[2:4])
	if err != nil {
		return rgba{}, err
	}
	decodedB, err := hex.DecodeString(trimmedH[4:6])
	if err != nil {
		return rgba{}, err
	}
	decodedA, err := hex.DecodeString(trimmedH[6:8])
	if err != nil {
		return rgba{}, err
	}
	return rgba{
		R: uint8(decodedR[0]),
		G: uint8(decodedG[0]),
		B: uint8(decodedB[0]),
		A: uint8(decodedA[0]),
	}, nil
}

// getRGBA is a helper function to convert a uint32 to an RGBA colour
func getRGBA(element uint32) rgba {
	colour := rgba{
		R: uint8(0xFF & (element >> 24)),
		G: uint8(0xFF & (element >> 16)),
		B: uint8(0xFF & (element >> 8)),
		// store the least significant bits as the alpha value
		A: uint8(0xFF & element),
	}
	colour.Hex = colour.printHex()
	return colour
}
