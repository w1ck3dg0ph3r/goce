package parsers

import (
	"fmt"
	"os"
	"regexp"
	"testing"
)

func TestAssembly(t *testing.T) {
	re := regexp.MustCompile(`^\t(\w+) \d+ \(([^:]+):(\d+)\)\t([^\t]+)\t(.*)`)
	m := re.FindSubmatch([]byte(`	0x0015 00021 (tmp/main.go:36)	MOVL	$100, BX`))
	for _, b := range m {
		fmt.Println(string(b))
	}
}

func TestParse(t *testing.T) {
	src, err := os.ReadFile("testdata/main.go")
	if err != nil {
		panic(err)
	}
	out, err := os.Open("testdata/buildoutput")
	if err != nil {
		panic(err)
	}
	res := &Result{}
	parseBuildOutput(res, src, out)
	if len(res.Assembly) == 0 {
		t.Fail()
	}
	if len(res.Mapping) == 0 {
		t.Fail()
	}
	if len(res.Diagnostics) == 0 {
		t.Fail()
	}
	for _, d := range res.Diagnostics {
		switch d := d.(type) {
		case HeapEscape:
			if d.Name != "" {
				fmt.Printf("heap escape: %s\n", d.Name)
			} else {
				fmt.Printf("heap escape: %s\n", d.Message)
			}
		}
	}
}
