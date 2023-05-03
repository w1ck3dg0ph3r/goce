package parsers

import (
	"strings"

	"github.com/w1ck3dg0ph3r/goce/compilers"
)

type Parser interface {
	Parse(output compilers.Result) Result
}

type Result struct {
	Assembly         string             `json:"assembly"`
	Mapping          []Mapping          `json:"mapping"`
	InliningAnalysis []InliningAnalysis `json:"inliningAnalysis"`
	HeapEscapes      []HeapEscape       `json:"heapEscapes"`
}

type Mapping struct {
	SourceLine    int `json:"source"`
	AssemblyStart int `json:"start"`
	AssemblyEnd   int `json:"end"`
}

type InliningAnalysis struct {
	Name      string   `json:"name"`
	Location  Location `json:"location"`
	CanInline bool     `json:"canInline"`
	Reason    string   `json:"reason"`
	Cost      int      `json:"cost"`
}

type HeapEscape struct {
	Name     string   `json:"name"`
	Location Location `json:"location"`
}

type Location struct {
	Line   int `json:"l"`
	Column int `json:"c"`
}

func FindMatching(output compilers.Result) Parser {
	switch {
	case strings.HasPrefix(output.CompilerInfo.Version, "1.20"):
		return currentParser{}
	default:
		return currentParser{}
	}
}
