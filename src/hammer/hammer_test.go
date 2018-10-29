package hammer

import (
	"testing"
)

var (
	prog = "qiime"
	path = "./otu-table.txt"
)

// test the otu table constructor
func TestConstructor(t *testing.T) {
	table, err := NewOTUtable(path, prog)
	if err != nil {
		t.Fatal(err)
	}
	if table.GetNumSamples() != 1 {
		t.Fatal("wrong number of samples collected from test file")
	}
	if table.GetTotalGenusOTUs() != 4 {
		t.Fatal("consensus lineage parsing not correct")
	}
}

// test the KeepTopN method
func TestTopN(t *testing.T) {
	table, _ := NewOTUtable(path, prog)
	if err := table.KeepTopN(3); err != nil {
		t.Fatal(err)
	}
	// make sure topN can't be called again
	if err := table.KeepTopN(3); err == nil {
		t.Fatal("at present, we only want to be able to call KeepTopN once")
	}
	// make sure the numOTUs check works
	table2, _ := NewOTUtable(path, prog)
	if err := table2.KeepTopN(5); err == nil {
		t.Fatal("n must be < len(otu table)")
	}
}
