// Copyright © 2018 Science and Technology Facilities Council (UK) <will.rowe@stfc.ac.uk>

package cmd

import (
	//"encoding/csv"
	"fmt"
	"log"
	"os"

	"github.com/pkg/profile"
	"github.com/spf13/cobra"
	"github.com/will-rowe/hulk/src/misc"
	"github.com/will-rowe/thor/src/colour"
	"github.com/will-rowe/thor/src/draw"
	"github.com/will-rowe/thor/src/hammer"
	"github.com/will-rowe/thor/src/version"
)

// the currently supported otu table formats
var supportedFormats [1]string = [1]string{"qiime"}

// the command line arguments
var (
	otuTables      *[]string // the input OTU tables
	format         *string   // the otuTable format
	colourSketches *string   // the reference colour sketches
	alphaAbundance *bool     // replace the alpha channel of the colour sketch with the OTU abundance
	storeTHORcsv   *bool     // also store the image as a csv of RGBA values
)

// hammerCmd represents the hammer command
var hammerCmd = &cobra.Command{
	Use:   "hammer",
	Short: "Hammer an OTU table into an image...",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		runHammer()
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return misc.CheckRequiredFlags(cmd.Flags())
	},
}

// a function to initialise the command line arguments
func init() {
	otuTables = hammerCmd.Flags().StringSliceP("otuTables", "i", []string{}, "input OTU table(s) to transform to hashed OTU RGBA images")
	format = hammerCmd.Flags().StringP("otuFormat", "f", "qiime", "the format of the input OTU table(s) (only QIIME currently supported")
	colourSketches = hammerCmd.Flags().StringP("colourSketches", "c", "", "the set of reference colour sketches (from `thor colour`)")
	alphaAbundance = hammerCmd.Flags().Bool("alphaAbundance", false, "include the OTU abundance (replaces existing alpha value of colour sketches")
	storeTHORcsv = hammerCmd.Flags().Bool("csv", false, "also store the image as a csv file of RGBA values")
	hammerCmd.MarkFlagRequired("otuTables")
	hammerCmd.MarkFlagRequired("colourSketches")
	hammerCmd.Flags().SortFlags = false
	RootCmd.AddCommand(hammerCmd)
}

// check the program input
func checkInput() error {
	// check specified format is supported
	var check bool
	for _, sf := range supportedFormats {
		if sf == *format {
			check = true
		}
	}
	if check == false {
		return fmt.Errorf("OTU table format not supported: %v", *format)
	}
	// check the OTU tables
	for _, otuTable := range *otuTables {
		if _, err := os.Stat(otuTable); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("file does not exist: %v", otuTable)
			} else {
				return fmt.Errorf("can't access file (check permissions): %v", otuTable)
			}
		}
		// TODO: check the supplied file is actually an OTU table

		// TODO: check that the supplied file is in the specified format

	}
	// check the colour sketch file
	if *colourSketches == "" {
		return fmt.Errorf("require --colourSketches, run `thor colour` if you haven't already")
	}
	if _, err := os.Stat(*colourSketches); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %v", *colourSketches)
		} else {
			return fmt.Errorf("can't access file (check permissions): %v", *colourSketches)
		}
	}
	return nil
}

/*
  The main function for the hammer subcommand
*/
func runHammer() {
	// set up profiling
	if *profiling == true {
		//defer profile.Start(profile.MemProfile, profile.ProfilePath("./")).Stop()
		defer profile.Start(profile.ProfilePath("./")).Stop()
	}
	// start logging
	logFH := misc.StartLogging((*outFile + ".log"))
	defer logFH.Close()
	log.SetOutput(logFH)
	log.Printf("thor (version %s)", version.VERSION)
	log.Printf("starting the hammer subcommand")
	// check the supplied files and then log some stuff
	log.Printf("checking parameters...")
	misc.ErrorCheck(checkInput())
	log.Printf("\tinput OTU tables:")
	for _, file := range *otuTables {
		log.Printf("\t\t%v", file)
	}
	log.Printf("\tOTU table format: %v", *format)
	log.Printf("\toutput file basename: %v", *outFile)
	log.Printf("\tinclude OTU abundance: %t", *alphaAbundance)
	log.Printf("\tstore CSV: %t", *storeTHORcsv)
	log.Printf("\tcolour sketches: %v", *colourSketches)
	// load the reference colour sketches
	css := make(colour.ColourSketchStore)
	misc.ErrorCheck(css.Load(*colourSketches))
	sketchLength := css.GetSketchLength()
	log.Printf("\tsketch length: %d", sketchLength)
	// process each OTU table
	log.Printf("processing OTU table(s)...")
	// TODO: should I make this run concurrently?
	for i, otuTable := range *otuTables {
		// read the OTU table
		table, err := hammer.NewOTUtable(otuTable, *format)
		misc.ErrorCheck(err)
		log.Printf("\ttable %d: %v", (i + 1), otuTable)
		log.Printf("\tnum. samples: %d", table.GetNumSamples())
		log.Printf("\tnum. OTU ids at genus level: %d", table.GetTotalGenusOTUs())
		// get the top N most abundant OTUs (and add padding if needed) for each sample
		misc.ErrorCheck(table.KeepTopN(sketchLength))
		// parse top OTUs, lookup the coloursketches and keep corresponding rgba slices for each sample
		sampleRGBAs, err := table.ColourTopN(css)
		misc.ErrorCheck(err)
		// process each sample, collecting the pixel vectors
		for j, sampleRGBA := range sampleRGBAs {
			// create the canvas
			img, err := draw.NewThorPNG(sketchLength, sketchLength)
			misc.ErrorCheck(err)
			// collect the pixel vectors
			for _, line := range sampleRGBA {
				if line == nil {
					line = colour.GetPadding(sketchLength)
				}
				err := img.DrawOTU(line)
				misc.ErrorCheck(err)
			}
			// write the png
			sample, err := table.GetSampleName(j)
			misc.ErrorCheck(err)
			filename := fmt.Sprintf("%v-%v.thor-image.png", *outFile, sample)
			misc.ErrorCheck(img.Save(filename))
		}
	}

}
