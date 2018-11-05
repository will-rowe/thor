// hammer contains the types/methods/functions to transform an OTU table into a rgba image

package hammer

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"image/color"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/will-rowe/thor/src/colour"
)

// otu
type otu struct {
	otu       string
	abundance int
}

// otuTable
type otuTable struct {
	program  string
	path     string
	comments [][]byte
	// the ordering of the outside slice of sampleNames, sampleData and topN are used to relate the data
	sampleNames [][]byte
	sampleData  []map[string]int
	topN        [][]otu
	totalOTUs   int
	// the COLOURSKETCH map
	ColourSketchStore colour.ColourSketchStore
}

// PrintComments returns the OTU table comments, formatted as a single string with newlines
func (otuTable *otuTable) PrintComments() string {
	var comments string
	for i := 0; i < len(otuTable.comments); i++ {
		// comments where stored in order they were read (newlines were also kept)
		comments = fmt.Sprintf("%v%v", comments, string(otuTable.comments[i]))
	}
	return comments
}

// GetNumSamples returns the number of samples found in the OTU table
func (otuTable *otuTable) GetNumSamples() int {
	return len(otuTable.sampleNames)
}

// GetSampleName returns the string formatted sample name, give the index position
func (otuTable *otuTable) GetSampleName(i int) (string, error) {
	if i > len(otuTable.sampleNames) {
		return "", fmt.Errorf("sample index position > number of samples!")
	}
	return string(otuTable.sampleNames[i]), nil
}

// GetTotalGenusOTUs returns the total number of genus level OTUs in the original OTU table file
func (otuTable *otuTable) GetTotalGenusOTUs() int {
	return otuTable.totalOTUs
}

// KeepTopN is a method to keep only the top N most abundant OTUs in each sample
// it clears the original sampleData and keeps the topN in a set of new slices
func (otuTable *otuTable) KeepTopN(n int) error {
	// make sure this method hasn't already been run
	if len(otuTable.topN[0]) != 0 {
		return fmt.Errorf("the KeepTopN method has already been run on this OTU table")
	}
	// make sure n > num OTUs in table
	if n > otuTable.GetTotalGenusOTUs() {
		return fmt.Errorf("requested number of top OTUs is greater than the total number of OTUs")
	}
	// sort each sample in a separate go routine and then update the top n otus
	var wg sync.WaitGroup
	for i := 0; i < len(otuTable.sampleData); i++ {
		wg.Add(1)
		go sortOTUs(otuTable, i, n, &wg)
	}
	wg.Wait()
	return nil
}

// ColourTopN returns the corresponding coloursketches for the TopN otus
// returns the sample ID, the slice of coloursketches and any error
func (otuTable *otuTable) ColourTopN() ([][][]color.RGBA, error) {
	rgbaLines := make([][][]color.RGBA, otuTable.GetNumSamples())
	// perform for each sample in the OTU table
	for i := range otuTable.sampleNames {
		rgbaLines[i] = make([][]color.RGBA, len(otuTable.topN[i]))
		// for each sample, range over the topN otus
		for j, otu := range otuTable.topN[i] {
			// skip padding lines
			if otu.otu == "padding" {
				continue
			}
			// lookup the otu in the css
			if _, ok := otuTable.ColourSketchStore[otu.otu]; !ok {
				// TODO: we've already checked that the topN OTUs are present in the REFSEQ db, this error should never happen here
				return nil, fmt.Errorf("sample %v: the genus name `%v` (abundance: %d) could not be found in the coloursketches", string(otuTable.sampleNames[i]), otu.otu, otu.abundance)
			} else {
				if rgba, err := otuTable.ColourSketchStore[otu.otu].PrintPNGline(); err != nil {
					return nil, err
				} else {
					rgbaLines[i][j] = rgba
				}
			}
		}
	}
	return rgbaLines, nil
}

// readQiimeTable will load a qiime file into the otuTable
func (otuTable *otuTable) readQiimeTable() error {
	// create a new reader
	fh, err := os.Open(otuTable.path)
	if err != nil {
		return err
	}
	defer fh.Close()
	r := bufio.NewReader(fh)
	// slurp off the comments
	for {
		char, err := r.Peek(1)
		if err != nil {
			return err
		}
		if char[0] == 35 {
			// either a comment or the header line
			line, err := r.ReadBytes('\n')
			if err != nil {
				return err
			}
			// add the comment and keep peeking
			if string(line[0:4]) != "#OTU" {
				otuTable.comments = append(otuTable.comments, line)
				// or finish the slurping, add the samples from the header and start reading lines
			} else {
				samples := bytes.Split(line, []byte("	"))
				numSamples := (len(samples) - 2)
				samples = samples[1 : 1+numSamples]
				otuTable.sampleNames = make([][]byte, numSamples)
				otuTable.sampleData = make([]map[string]int, numSamples)
				otuTable.topN = make([][]otu, numSamples)
				for i, sample := range samples {
					otuTable.sampleNames[i] = sample
					otuTable.sampleData[i] = make(map[string]int)
				}
				break
			}
		} else {
			return fmt.Errorf("no comment or header lines found in Qiime OTU table")
		}
	}
	var counter int
	tsvReader := csv.NewReader(r)
	tsvReader.Comma = '	'
	for {
		line, err := tsvReader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		// grab the consensus lineage, check for genus and keep it
		consensusLineage := strings.Split(line[len(line)-1], ";g__")
		if len(consensusLineage) != 2 {
			continue
		}
		if consensusLineage[1] == "" {
			continue
		}
		// add the abundance values to the corresponding samples
		for i := 1; i <= len(otuTable.sampleData); i++ {
			value, err := strconv.Atoi(line[i])
			if err != nil {
				return err
			}
			if _, ok := otuTable.sampleData[i-1][consensusLineage[1]]; !ok {
				otuTable.sampleData[i-1][consensusLineage[1]] = value
			} else {
				otuTable.sampleData[i-1][consensusLineage[1]] += value
			}
		}
		counter++
	}
	otuTable.totalOTUs = counter
	return nil
}

// NewOTUtable is the otuTable constructor
func NewOTUtable(path, prog string) (*otuTable, error) {
	table := &otuTable{
		path:    path,
		program: prog,
	}
	// read in the file
	var err error
	switch table.program {
	case "qiime":
		err = table.readQiimeTable()
	default:
		err = fmt.Errorf("unsupported OTU table format: %v", prog)
	}
	if err != nil {
		return nil, err
	}
	return table, nil
}

// sortOTUs is a function to sort the OTUs by decreasing abundance, keeping only the top N
func sortOTUs(otuTable *otuTable, sampleID, n int, wg *sync.WaitGroup) {
	defer wg.Done()
	var topNotus []otu
	// put the otus into a slice
	for k, v := range otuTable.sampleData[sampleID] {
		// don't include OTUs that aren't in the REFSEQ database
		// TODO: this is a bit cludgy, I'll make it better...
		//if _, ok := otuTable.ColourSketchStore[k]; !ok {
		//	continue
		//}
		topNotus = append(topNotus, otu{k, v})
	}
	// sort
	sort.Slice(topNotus, func(i, j int) bool {
		return topNotus[i].abundance > topNotus[j].abundance
	})
	// update the OTUtable with the top n otus
	otuTable.topN[sampleID] = topNotus[0:n]
	// remove the full map for this sample
	otuTable.sampleData[sampleID] = make(map[string]int)
	// update any 0 abundance included in the topN to be marked as padding
	for i := 0; i < n; i++ {
		if otuTable.topN[sampleID][i].abundance == 0 {
			otuTable.topN[sampleID][i].otu = "padding"
		}
	}
	return
}
