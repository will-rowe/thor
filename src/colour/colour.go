//
package colour

import (
	"encoding/hex"
	"errors"
	"fmt"
)

const (
	ALPHA        = 255
)

// colourSketch is a struct to hold a colour encoded sketch
type colourSketch []*rgba

// Print is a method to print the coloured histoSketch as a csv line (either in rgb or hex)
func (colourSketch *colourSketch) Print(hex bool) (string, error) {
	var line string
	for _, value := range *colourSketch {
		// make sure the rgb is okay to use
		if err := value.checker(); err != nil {
			return "", err
		}
		if hex {
			line += value.hex + ","
		} else {
			line += value.printRGBA() + ","
		}
	}
	return line, nil
}

// rgb is a struct to hold the colour information for one histosketch element
type rgba struct {
	r uint8
	g uint8
	b uint8
	a uint8
	hex string
}

// checker is a method to check the rgb can be used
func (rgba *rgba) checker() error {
	if rgba.hex == "" {
		return errors.New("this is an uninitialised rgb")
	}
	return nil
}

// printRGB is a method to convert an rbga struct to a rgba string
func (rgba *rgba) printRGBA() string {
	return fmt.Sprintf("rgba(%d,%d,%d,%d)", rgba.r, rgba.g, rgba.b, rgba.a)
}

// printHex is a method to convert an rgba struct to a hex string
func (rgba *rgba) printHex() string {
	return fmt.Sprintf("#%02X%02X%02X%02X", rgba.r, rgba.g, rgba.b, rgba.a)
}

// ColourHistosketch is is the colourSketch constructor
func ColourHistosketch(histoSketch []uint32) colourSketch {
	// prepare the coloured histosketch
	colours := make(colourSketch, len(histoSketch))
	// iterate over the values in the histosketch
	for i := 0; i < len(histoSketch); i++ {
		colours[i] = getRGBA(histoSketch[i])
	}
	return colours
}

// getRGBA is a helper function to convert a uint32 to an RGBA colour
func getRGBA(element uint32) *rgba {
	colourSketch := &rgba{
		r: uint8(0xFF & (element >> 24)),
		g: uint8(0xFF & (element >> 16)),
		b: uint8(0xFF & (element >> 8)),
		// store the least significant bits as the alpha value
		a: uint8(0xFF & element),
	}
	colourSketch.hex = colourSketch.printHex()
	return colourSketch
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
		r: uint8(decodedR[0]),
		g: uint8(decodedG[0]),
		b: uint8(decodedB[0]),
		a: uint8(decodedA[0]),
	}, nil
}
