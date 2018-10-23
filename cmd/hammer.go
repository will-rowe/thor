// Copyright © 2018 Science and Technology Facilities Council (UK) <will.rowe@stfc.ac.uk>

package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/pkg/profile"
	"github.com/will-rowe/hulk/src/misc"
	"github.com/will-rowe/thor/src/colour"
	"github.com/will-rowe/thor/src/version"
)

// the command line arguments
var (
	otuTables      *[]string // the input OTU tables
	colourSketches *string   // the reference colour sketches
	alphaAbundance	*bool	// replace the alpha channel of the colour sketch with the OTU abundance
	storePng       *bool     // also store the image as a png
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
	otuTables = hammerCmd.Flags().StringSliceP("otuTables", "i", []string{}, "input OTU tables to transform to hashed OTU rgb images")
	colourSketches = hammerCmd.Flags().StringP("colourSketches", "c", "", "the set of reference colour sketches (from `thor colour`)")
	alphaAbundance = hammerCmd.Flags().Bool("alphaAbundance", false, "include the OTU abundance (replaces existing alpha value of colour sketches")
	storePng = hammerCmd.Flags().Bool("png", false, "also store the image as a png file")
	hammerCmd.MarkFlagRequired("otuTables")
	hammerCmd.MarkFlagRequired("colourSketches")
	RootCmd.AddCommand(hammerCmd)
}

// check the program input
func checkInput() error {
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
	log.Printf("\tinput files:")
	for _, file := range *otuTables {
		log.Printf("\t\t%v", file)
	}
	log.Printf("\tcolour sketches:")
	log.Printf("\t\t%v", *colourSketches)
	log.Printf("\toutput file basename: %v", *outFile)
	log.Printf("\tinclude OTU abundance: %t", *alphaAbundance)
	log.Printf("\tstore PNG: %t", *storePng)
	// load the reference colour sketches
	css := make(colour.ColourSketchStore)
	misc.ErrorCheck(css.Load(*colourSketches))
	// process each OTU table
	for i, otuTable := range *otuTables {
		fmt.Println(i, otuTable)

		// set up the hammer object


		// parse each OTU and grab the corresponding colour sketch


		// if we are overwriting the alpha channel, replace this now with the OTU abundance


		// add the colour sketch to the hammer object


		// start hammering
		// only keep top X most abundant OTUs, where X = len(colour sketch), giving us a square image

		// send the final image on
		

	}


	// save all images (and PNGs if requested)
}
