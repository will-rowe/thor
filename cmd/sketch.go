// Copyright © 2018 Science and Technology Facilities Council (UK) <will.rowe@stfc.ac.uk>

package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkg/profile"
	"github.com/spf13/cobra"
	"github.com/will-rowe/hulk/src/histosketch"
	"github.com/will-rowe/hulk/src/misc"
	"github.com/will-rowe/hulk/src/stream"
	hVersion "github.com/will-rowe/hulk/src/version"
	tVersion "github.com/will-rowe/thor/src/version"
)

// the command line arguments
var (
	fasta      *string  //	FASTA file(s) to sketch, will perform a glob using the given string
	inputSeqs  []string // the input files are put in this slice once the --fasta CL option is parsed
	sketchAlgo *string  // the sketching algorithm to use (histosketch or minhash)
	kSize      *int     // size of k-mer
	epsilon    *float64  // epsilon value for countminsketch generation
	delta      *float64  // delta value for countminsketch generation
	minCount   *int     // minimum count number for a kmer to be added to the histosketch from this interval
	sketchSize *uint    // size of sketch
)

// the sketchCmd
var sketchCmd = &cobra.Command{
	Use:   "sketch",
	Short: "Create a set of coloured histosketches from a set of FASTA files",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		runSketch()
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return misc.CheckRequiredFlags(cmd.Flags())
	},
}

// a function to initialise the command line arguments
func init() {
	fasta = sketchCmd.Flags().StringP("fasta", "f", "", "FASTA file(s) to sketch (can also pipe STDIN)")
	sketchAlgo = sketchCmd.Flags().StringP("sketchAlgo", "a", "histosketch", "the sketching algorithm to use (histosketch or minhash)")
	kSize = sketchCmd.Flags().IntP("kmerSize", "k", 21, "size of k-mer")
	epsilon = sketchCmd.Flags().Float64P("epsilon", "e", 0.00001, "epsilon value for countminsketch generation")
	delta = sketchCmd.Flags().Float64P("delta", "d", 0.90, "delta value for countminsketch generation")
	minCount = sketchCmd.Flags().IntP("minCount", "m", 1, "minimum k-mer count for it to be histosketched for a given interval")
	sketchSize = sketchCmd.Flags().UintP("sketchSize", "s", 100, "size of sketch")
	sketchCmd.Flags().SortFlags = false
	RootCmd.AddCommand(sketchCmd)
}

//  a function to check user supplied parameters
func sketchParamCheck() error {
	// check the algorithm is minhash or histosketch
	switch *sketchAlgo {
	case "histosketch":
		log.Printf("\tsketching algorithm: histosketch")
	case "minhash":
		log.Printf("\tsketching algorithm: minhash")
	default:
		fmt.Println("--sketchAlgo must be either histosketch or minhash")
		return fmt.Errorf("--sketchAlgo must be either histosketch or minhash")
	}
	// check if using STDIN or file(s)
	if *fasta == "" {
		stat, err := os.Stdin.Stat()
		if err != nil {
			fmt.Println("error with STDIN")
			return fmt.Errorf("error with STDIN")
		}
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			fmt.Println("no STDIN found")
			return fmt.Errorf("no STDIN found")
		}
		log.Printf("\tinput file: using STDIN")
		return nil
	}
	// check the supplied file(s)
	return checkInputFiles()
}

// if files are being read, check they exist and are FASTQ/FASTA
func checkInputFiles() error {
	var err error
	inputSeqs, err = filepath.Glob(*fasta)
	misc.ErrorCheck(err)
	for _, fastqFile := range inputSeqs {
		if _, err := os.Stat(fastqFile); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("file does not exist: %v", fastqFile)
			} else {
				return fmt.Errorf("can't access file (check permissions): %v", fastqFile)
			}
		}
		suffix1, suffix2, suffix3 := "fasta", "fna", "fa"
		splitFilename := strings.Split(fastqFile, ".")
		var ext string
		if splitFilename[len(splitFilename)-1] == "gz" {
			ext = splitFilename[len(splitFilename)-2]
		} else {
			ext = splitFilename[len(splitFilename)-1]
		}
		switch ext {
		case suffix1:
			continue
		case suffix2:
			continue
		case suffix3:
			continue
		case "":
			return fmt.Errorf("could not parse filename")
		default:
			return fmt.Errorf("does not look like a %v file: %v", suffix1, fastqFile)
		}
	}
	return nil
}

/*
  The main function for the sketch subcommand
*/
func runSketch() {
	// set up profiling
	if *profiling == true {
		//defer profile.Start(profile.MemProfile, profile.ProfilePath("./")).Stop()
		defer profile.Start(profile.ProfilePath("./")).Stop()
	}
	// start logging
	logFH := misc.StartLogging((*outFile + ".log"))
	defer logFH.Close()
	log.SetOutput(logFH)
	log.Printf("thor (version %s)", tVersion.VERSION)
	log.Printf("\tuses: hulk (version %s)", hVersion.VERSION)
	log.Printf("starting the sketch subcommand")
	// check the supplied files and then log some stuff
	log.Printf("checking parameters...")
	misc.ErrorCheck(sketchParamCheck())
	log.Printf("\tinput files:")
	for _, file := range inputSeqs {
		log.Printf("\t\t%v", file)
	}
	log.Printf("\toutput file basename: %v", *outFile)
	log.Printf("\tno. processors: %d", *proc)
	log.Printf("\tk-mer size: %d", *kSize)
	log.Printf("\tmin. k-mer count: %d", *minCount)
	log.Printf("\tsketch size: %d", *sketchSize)
	// create the base countmin sketch for recording the k-mer spectrum
	log.Printf("creating the base countmin sketch for kmer counting...")
	// TODO: epsilon and delta values need some checking
	spectrum := histosketch.NewCountMinSketch(*epsilon, *delta, 1.0)
	log.Printf("\tnumber of tables: %d", spectrum.Tables())
	log.Printf("\tnumber of counters per table: %d", spectrum.Counters())
	log.Printf("sketching %d files...", len(inputSeqs))

	var wg sync.WaitGroup
	wg.Add(len(inputSeqs))
	for i := 0; i < len(inputSeqs); i++ {
		go func(file string) {
			defer wg.Done()
			// set up output files for this sequence
			fName := strings.Split(file, "/")
			sketchFile := *outFile + "-hulk." + fName[len(fName)-1] + ".sketch"
			// create the pipeline
			pipeline := stream.NewPipeline()
			// initialise processes
			dataStream := stream.NewDataStreamer()
			fastqHandler := stream.NewFastqHandler()
			fastqChecker := stream.NewFastqChecker()
			counter := stream.NewCounter()
			sketcher := stream.NewSketcher()
			// add in the process parameters TODO: consolidate and remove some of these
			dataStream.InputFile = []string{file}
			fastqHandler.Fasta, counter.Fasta = true, true
			fastqChecker.Ksize, counter.Ksize = *kSize, *kSize
			counter.Interval = 0
			counter.Spectrum, sketcher.Spectrum = spectrum.Copy(), spectrum.Copy()
			counter.NumCPU, sketcher.NumCPU = *proc, *proc
			counter.SketchSize, sketcher.SketchSize = *sketchSize, *sketchSize
			counter.ChunkSize = -1
			sketcher.MinCount = float64(*minCount)
			sketcher.DecayRatio = 1.0
			sketcher.OutFile = sketchFile
			// arrange pipeline processes
			fastqHandler.Input = dataStream.Output
			fastqChecker.Input = fastqHandler.Output
			counter.Input = fastqChecker.Output
			sketcher.Input = counter.TheCollector
			// submit each process to the pipeline to be run
			pipeline.AddProcesses(dataStream, fastqHandler, fastqChecker, counter, sketcher)
			pipeline.Run()
		}(inputSeqs[i])
	}
	wg.Wait()
	log.Printf("finished")
}
