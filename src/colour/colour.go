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
type colourSketch []*rgb

// Print is a method to print the coloured histoSketch as a csv line (either in rgb or hex)
func (colourSketch *colourSketch) Print(hex bool) (string, error) {
	var line string
	for _, value := range *colourSketch {
		// make sure the rgb is okay to use
		if err := value.checker(); err != nil {
			return "", err
		}
		if hex {
			line += value.printHex() + ","
		} else {
			line += value.printRGB() + ","
		}
	}
	return line, nil
}

// rgb is a struct to hold the colour information for one histosketch element
type rgb struct {
	r uint8
	g uint8
	b uint8
	a uint8
}

// checker is a method to check the rgb can be used
func (rgb *rgb) checker() error {
	if (rgb.r + rgb.g + rgb.b) == 0 {
		return errors.New("this is an uninitialised rgb")
	}
	return nil
}

// printRGB is a method to convert an rbg struct to a string
func (rgb *rgb) printRGB() string {
	return fmt.Sprintf("rgb(%d,%d,%d)", rgb.r, rgb.g, rgb.b)
}

// printHex is a method to convert an rgb struct into a hex string
func (rgb *rgb) printHex() string {
	return fmt.Sprintf("#%02X%02X%02X", rgb.r, rgb.g, rgb.b)
}

// ColourHistosketch is is the colourSketch constructor
func ColourHistosketch(histoSketch []uint64) colourSketch {
	// prepare the coloured histosketch
	colours := make(colourSketch, len(histoSketch))
	// iterate over the values in the histosketch
	for i := 0; i < len(histoSketch); i++ {
		colours[i] = getRGB(histoSketch[i])
	}
	return colours
}

// getRGB is a helper function to convert an int64 to an RGB colour
func getRGB(element uint64) *rgb {
	return &rgb{
		r: uint8((element & 0xFF0000) >> 16),
		g: uint8((element & 0x00FF00) >> 8),
		b: uint8((element & 0x0000FF)),
		//TODO: could add option to use the histosketches weights to adjust the alpha?
		a: ALPHA,
	}
}

// hex2rgb is a helper function to convert a hex string back to an RGB struct
func hex2rgb(h string) (*rgb, error) {
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
	return &rgb{
		r: uint8(decodedR[0]),
		g: uint8(decodedG[0]),
		b: uint8(decodedB[0]),
	}, nil
}
