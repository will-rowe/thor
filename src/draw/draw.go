package draw

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

// PAD_COLOUR is the colour to use to fill the rest of the PNG if not enough OTU vectors are given
var PAD_COLOUR = color.RGBA{255, 255, 255, 255}

// thorPNG
type thorPNG struct {
	canvas   *image.RGBA
	xy       int
	padding  int
	currentY int
}

// GetPadding is a method to return the number of padding rows needed to square the PNG
func (thorPNG *thorPNG) GetPadding() int {
	return thorPNG.padding
}

// DrawOTU method will add a row of pixels to the PNG,
func (thorPNG *thorPNG) DrawOTU(colours []color.RGBA) error {
	// check the incoming vector is compatible with the image
	if len(colours) != thorPNG.xy {
		return fmt.Errorf("was expecting sketch of length %d, received vector of length %d", thorPNG.xy, len(colours))
	}
	// check if the image is full yet
	if (thorPNG.currentY + thorPNG.padding) == len(colours) {
		return fmt.Errorf("image full")
	}
	// add each pixel to the new row in the image
	for x := 0; x < thorPNG.xy; x++ {
		thorPNG.canvas.Set(x, thorPNG.currentY, colours[x])
	}
	thorPNG.currentY++
	return nil
}

// Save method will check and save the thorPNG to disk
func (thorPNG *thorPNG) Save(filepath string, padding bool) error {
	// check the PNG has been built from enough OTUs for current canvas size
	if thorPNG.xy != thorPNG.currentY {
		// add padding to the end of the PNG if requested
		if padding {
			for thorPNG.currentY < thorPNG.xy {
				for x := 0; x < thorPNG.xy; x++ {
					thorPNG.canvas.Set(x, thorPNG.currentY, PAD_COLOUR)
				}
				thorPNG.currentY++
			}
		} else {
			// TODO: if no padding requested, remove the empty rows from the canvas

		}
	}
	fh, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer fh.Close()
	// encode as png and save (returning any error)
	return png.Encode(fh, thorPNG.canvas)

}

// NewThorPNG is the thorPNG constructor
func NewThorPNG(sketchLength, numOtus int) (*thorPNG, error) {
	// we adding padding if sketchLength < number of otus
	var pad int
	if sketchLength >= numOtus {
		pad = sketchLength - numOtus
	} else {
		return nil, fmt.Errorf("number of OTUs > sketch length (%d : %d). Suggest supplying top %d OTUs.", numOtus, sketchLength, sketchLength)
	}
	// create the canvas so that it is a square, equal to the sketchLength
	return &thorPNG{
		canvas:   image.NewRGBA(image.Rect(0, 0, sketchLength, sketchLength)),
		xy:       sketchLength,
		padding:  pad,
		currentY: 0,
	}, nil
}
