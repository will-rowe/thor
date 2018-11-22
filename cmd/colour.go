// Copyright © 2018 Science and Technology Facilities Council (UK) <will.rowe@stfc.ac.uk>

package cmd

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/will-rowe/hulk/src/histosketch"
	"github.com/will-rowe/hulk/src/misc"
	"github.com/will-rowe/thor/src/colour"
	"github.com/will-rowe/thor/src/hammer"
)

// the command line arguments
var (
	sketchDir *string // the directory containing the sketches
	recursive *bool   // recursively search the supplied directory
	storeCSV  *bool   // also write the colour sketches to a plain text csv file
)

// the sketches
var hSketches map[string]*histosketch.SketchStore

// colourCmd represents the colour command
var colourCmd = &cobra.Command{
	Use:   "colour",
	Short: "Colour a reference set of sketches",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		runColour()
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return misc.CheckRequiredFlags(cmd.Flags())
	},
}

// a function to initialise the command line arguments
func init() {
	sketchDir = colourCmd.Flags().StringP("sketchDir", "d", "./", "the directory containing the sketches to smash (compare)...")
	recursive = colourCmd.Flags().Bool("recursive", false, "recursively search the supplied sketch directory (-d)")
	storeCSV = colourCmd.Flags().Bool("storeCSV", false, "also write the colour sketches (as hex) to a plain text csv file")
	colourCmd.Flags().SortFlags = false
	RootCmd.AddCommand(colourCmd)
}

// makeColourSketches will colour the sketches and then write to a THOR data structure (and csv if requested)
func makeColourSketches() error {
	// create the csv outfile if asked for
	var csvWriter *csv.Writer
	if *storeCSV {
		csvFile, err := os.Create((*outFile + "-coloursketches.csv"))
		defer csvFile.Close()
		if err != nil {
			return err
		}
		csvWriter = csv.NewWriter(csvFile)
		defer csvWriter.Flush()
	}
	// create an ordering
	ordering := make([]string, len(hSketches))
	count := 0
	for id := range hSketches {
		ordering[count] = id
		count++
	}
	sort.Strings(ordering)
	var wg sync.WaitGroup
	csc := colour.NewColourSketchChan()
	// set up colour sketch store
	css := make(colour.ColourSketchStore)
	// colour the sketches
	for _, id := range ordering {
		wg.Add(1)
		// get the sketch values and launch go routines
		go func(sketch []uint, id string) {
			defer wg.Done()
			// convert sketch values to uint32, send error if they overflow
			values := make([]uint32, len(sketch))
			var err bool
			for i := 0; i < len(sketch); i++ {
				if sketch[i] > math.MaxUint32 {
					err = true
					break
				}
				values[i] = uint32(sketch[i])
			}
			// if the sketch fits, colour and send it
			if !err {
				csc.Send(colour.NewColourSketch(values, id), nil)
			} else {
				csc.Send(nil, fmt.Errorf("sketch contains element that overflows uint32: %v", id))
			}

		}(hSketches[id].Sketch, id)
	}
	go func() {
		wg.Wait()
		close(csc)
	}()

	// collect the coloursketches
	for parcel := range csc {
		// check for errors
		coloursketch, err := parcel.Unpack()
		misc.ErrorCheck(err)
		// clean up the id so that only the genus remains
		tmp1 := strings.TrimSuffix(coloursketch.Id, ".sketch")
		tmp2 := strings.Split(tmp1, "/")
		if len(tmp2) == 1 {
			coloursketch.Id = tmp2[0]
		} else {
			coloursketch.Id = tmp2[len(tmp2)-1]
		}
		// add this coloursketch to the store
		if _, ok := css[coloursketch.Id]; !ok {
			css[coloursketch.Id] = coloursketch
		} else {
			return fmt.Errorf("duplicate sketch name found: %v", coloursketch.Id)
		}
		// write this colour sketch (in hex) to the csv
		if *storeCSV {
			colours, err := coloursketch.PrintCSVline(true)
			if err != nil {
				return err
			}
			if err := csvWriter.Write([]string{coloursketch.Id, colours}); err != nil {
				return err
			}
		}
	}
	// add a padding line (slice of 0s) to the store
	padLine := make([]uint32, css.GetSketchLength())
	css[hammer.PAD_LINE] = colour.NewColourSketch(padLine, hammer.PAD_LINE)
	// encode and write the colour sketch map to disk
	return css.Dump(*outFile + "-coloursketches.thor")
}

/*
  The main function for the colour subcommand
*/
func runColour() {
	// add a slash if not already present in dir param
	sDir := []byte(*sketchDir)
	if sDir[len(sDir)-1] != 47 {
		sDir = append(sDir, 47)
	}
	// create the sketch pile
	var err error
	hSketches, _, err = histosketch.CreateSketchCollection(string(sDir), *recursive)
	misc.ErrorCheck(err)
	// check we have at least 2 sketches
	if len(hSketches) < 1 {
		fmt.Println("need at least 1 sketch!")
		os.Exit(1)
	}
	// colour the sketches
	misc.ErrorCheck(makeColourSketches())
}
