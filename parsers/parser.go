package parsers

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/w1ck3dg0ph3r/goce/compilers"
)

type Parser interface {
	Parse(output compilers.Result) Result
}

type Result struct {
	Assembly         string             `json:"assembly"`
	Mapping          []Mapping          `json:"mapping"`
	InliningAnalysis []InliningAnalysis `json:"inliningAnalysis"`
	InlinedCalls     []InlinedCall      `json:"inlinedCalls"`
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

type InlinedCall struct {
	Name     string   `json:"name"`
	Location Location `json:"location"`
	Length   int      `json:"length"`
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
	ver, err := semver.NewVersion(output.CompilerInfo.Version)
	if err != nil {
		panic(fmt.Sprintf("incorrect compiler version: %s", output.CompilerInfo.Version))
	}
	lastSupported := semver.MustParse("1.18")
	switch {
	case ver.GreaterThan(lastSupported):
		return currentParser{}
	default:
		return nil
	}
}
