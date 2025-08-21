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
	BuildOutput string        `json:"buildOutput"`
	Assembly    string        `json:"assembly"`
	Mapping     []Mapping     `json:"mapping"`
	Diagnostics []IDiagnostic `json:"diagnostics"`
}

type Mapping struct {
	SourceLine    int `json:"source"`
	AssemblyStart int `json:"start"`
	AssemblyEnd   int `json:"end"`
}

// Can be one of:
// [InliningAnalysis], [InlinedCall], [HeapEscape]
type IDiagnostic any

type Diagnostic struct {
	Type  DiagnosticType `json:"type"`
	Range Range          `json:"range"`
}

type DiagnosticType string

const (
	DiagnosticInliningAnalysis DiagnosticType = "inliningAnalysis"
	DiagnosticInlinedCall      DiagnosticType = "inlinedCall"
	DiagnosticHeapEscape       DiagnosticType = "heapEscape"
)

type Range struct {
	Start Location `json:"s"`
	End   Location `json:"e"`
}

type Location struct {
	Line   int `json:"l"`
	Column int `json:"c"`
}

type InliningAnalysis struct {
	Diagnostic
	Name      string `json:"name"`
	CanInline bool   `json:"canInline"`
	Reason    string `json:"reason"`
	Cost      int    `json:"cost"`
}

type InlinedCall struct {
	Diagnostic
	Name string `json:"name"`
}

type HeapEscape struct {
	Diagnostic
	Name    string `json:"name"`
	Message string `json:"message"`
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
