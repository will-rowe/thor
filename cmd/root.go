// Copyright © 2018 Science and Technology Facilities Council (UK) <will.rowe@stfc.ac.uk>

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// the command line arguments
var (
	proc           *int                                                      // number of processors to use
	outFile        *string                                                   // basename for the outfile(s)
	defaultOutFile = "./thor-" + string(time.Now().Format("20060102150405")) // a default output filename
	profiling      *bool                                                     // create profile for go pprof
	defaultLogFile = "./thor.log"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "thor",
	Short: "Transforming Hashed Otus to Rgb",
	Long: `
THOR is a tool that...

It works by ...`,
}

/*
  A function to add all child commands to the root command and sets flags appropriately
*/
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// check the persistent flags
	if err := persChecker(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// launch subcommand
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// init is a function to initialise the persistent flags
func init() {
	proc = RootCmd.PersistentFlags().IntP("processors", "p", 1, "number of processors to use")
	outFile = RootCmd.PersistentFlags().StringP("outFile", "o", defaultOutFile, "directory and basename for saving the outfile(s)")
	profiling = RootCmd.PersistentFlags().Bool("profiling", false, "create the files needed to profile THOR using the go tool pprof")
}

// persChecker checks the persistent flags
func persChecker() error {
	// setup the outFile
	filePath := filepath.Dir(*outFile)
	if filePath != "." {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			if err := os.MkdirAll(filePath, 0700); err != nil {
				return fmt.Errorf("can't create specified output directory: %v", err)
			}
		}
	}
	// set number of processors to use
	if *proc <= 0 || *proc > runtime.NumCPU() {
		*proc = runtime.NumCPU()
	}
	runtime.GOMAXPROCS(*proc)
	return nil
}
