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
type ColourSketchStore map[string]colourSketch

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

// colourSketch is a struct to hold a colour encoded sketch
type colourSketch []*rgba

// Print is a method to print the coloured sketch as a csv line (either in rgb or hex)
func (colourSketch *colourSketch) Print(printHex bool) (string, error) {
	var line string
	for _, value := range *colourSketch {
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
func NewColourSketch(sketch []uint32) colourSketch {
	// prepare the coloured sketch
	colours := make(colourSketch, len(sketch))
	// iterate over the values in the sketch and convert them to rgba
	for i := 0; i < len(sketch); i++ {
		colours[i] = getRGBA(sketch[i])
	}
	return colours
}

// getRGBA is a helper function to convert a uint32 to an RGBA colour
func getRGBA(element uint32) *rgba {
	colourSketch := rgba{
		R: uint8(0xFF & (element >> 24)),
		G: uint8(0xFF & (element >> 16)),
		B: uint8(0xFF & (element >> 8)),
		// store the least significant bits as the alpha value
		A: uint8(0xFF & element),
	}
	colourSketch.Hex = colourSketch.printHex()
	return &colourSketch
}

// hex2rgba is a helper function to convert a hex string back to an rgba struct
func hex2rgba(h string) (*rgba, error) {
	trimmedH := h
	if trimmedH[0] == '#' {
		trimmedH = h[1:]
	}
	if len(trimmedH) != 6 {
		return nil, errors.New(fmt.Sprintf("Invalid hex string: %s", h))
	}
	decodedR, err := hex.DecodeString(trimmedH[0:2])
	if err != nil {
		return nil, err
	}
	decodedG, err := hex.DecodeString(trimmedH[2:4])
	if err != nil {
		return nil, err
	}
	decodedB, err := hex.DecodeString(trimmedH[4:6])
	if err != nil {
		return nil, err
	}
	decodedA, err := hex.DecodeString(trimmedH[6:8])
	if err != nil {
		return nil, err
	}
	return &rgba{
		R: uint8(decodedR[0]),
		G: uint8(decodedG[0]),
		B: uint8(decodedB[0]),
		A: uint8(decodedA[0]),
	}, nil
}
