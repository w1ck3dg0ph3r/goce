package parsers

import (
	"bufio"
	"bytes"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/w1ck3dg0ph3r/goce/compilers"
)

type currentParser struct{}

func (currentParser) Parse(output compilers.Result) Result {
	var res Result
	parseBuildOutput(&res, output.BuildOutput)
	parseObjdumpOutput(&res, output.ObjdumpOutput)
	return res
}

func parseBuildOutput(res *Result, output io.Reader) {
	sc := bufio.NewScanner(output)

	mainFilenameBytes := []byte("./main.go")

	for sc.Scan() {
		if err := sc.Err(); err != nil {
			break
		}

		var match [][]byte

		// Inlining
		if match = reCanInline.FindSubmatch(sc.Bytes()); match != nil {
			if !bytes.Equal(match[reCanInlineFile], mainFilenameBytes) {
				continue
			}
			fc := InliningAnalysis{
				Name: string(match[reCanInlineName]),
				Location: Location{
					Line:   mustParseInt(match[reCanInlineLine]),
					Column: mustParseInt(match[reCanInlineColumn]),
				},
				CanInline: true,
			}
			cost, _ := strconv.Atoi(string(match[reCanInlineCost]))
			fc.Cost = cost
			res.InliningAnalysis = append(res.InliningAnalysis, fc)
		}
		if match = reCannotInline.FindSubmatch(sc.Bytes()); match != nil {
			if !bytes.Equal(match[reCannotInlineFile], mainFilenameBytes) {
				continue
			}
			fc := InliningAnalysis{
				Name: string(match[reCannotInlineName]),
				Location: Location{
					Line:   mustParseInt(match[reCannotInlineLine]),
					Column: mustParseInt(match[reCannotInlineColumn]),
				},
				CanInline: false,
				Reason:    string(match[reCannotInlineReason]),
			}
			res.InliningAnalysis = append(res.InliningAnalysis, fc)
		}

		// Heap escapes
		if match = reEscapesToHeap.FindSubmatch(sc.Bytes()); match != nil {
			if !bytes.Equal(match[reEscapesToHeapFile], mainFilenameBytes) {
				continue
			}
			he := HeapEscape{
				Name: string(match[reCannotInlineName]),
				Location: Location{
					Line:   mustParseInt(match[reEscapesToHeapLine]),
					Column: mustParseInt(match[reEscapesToHeapColumn]),
				},
			}
			res.HeapEscapes = append(res.HeapEscapes, he)
		}
	}
}

func parseObjdumpOutput(res *Result, output io.Reader) {
	sc := bufio.NewScanner(output)

	assembly := strings.Builder{}
	mainFilenameBytes := []byte("main.go")
	lastSourceLine := 0
	assemblyLine := 0

	for sc.Scan() {
		if err := sc.Err(); err != nil {
			break
		}

		var match [][]byte

		if match = reAssemblyText.FindSubmatch(sc.Bytes()); match != nil {
			assembly.Write([]byte(sc.Text()))
			assembly.WriteRune('\n')
			assemblyLine++
		}

		if match = reAssembly.FindSubmatch(sc.Bytes()); match != nil {
			assembly.Write(match[reAssemblyAddress])
			assembly.WriteRune('\t')
			assembly.Write(match[reAssemblyCode])
			assembly.WriteRune('\n')
			assemblyLine++
			if bytes.Equal(match[reAssemblyFile], mainFilenameBytes) {
				lineNumber, _ := strconv.Atoi(string(match[reAssemblyLine]))
				if lineNumber > lastSourceLine {
					res.Mapping = append(res.Mapping, Mapping{
						SourceLine:    lineNumber,
						AssemblyStart: assemblyLine,
						AssemblyEnd:   assemblyLine,
					})
				} else {
					lastMapping := &res.Mapping[len(res.Mapping)-1]
					lastMapping.AssemblyEnd = assemblyLine
				}
				lastSourceLine = lineNumber
			}
		}
	}

	res.Assembly = assembly.String()
}

func mustParseInt(s []byte) int {
	i, err := strconv.Atoi(string(s))
	if err != nil {
		panic(err)
	}
	return i
}

var reAssemblyText = regexp.MustCompile(`^TEXT`)

var reAssembly = regexp.MustCompile(`^\s+(.+?):(\d+)\t+(\w+)\t+([^\t]+)\t+(.*?)\t*$`)

const (
	reAssemblyFile = iota + 1
	reAssemblyLine
	reAssemblyAddress
	reAssemblyOpcodes
	reAssemblyCode
)

var reCanInline = regexp.MustCompile(`^(.+?):(\d+):(\d+): can inline (\w+) with cost (\d+)`)

const (
	reCanInlineFile = iota + 1
	reCanInlineLine
	reCanInlineColumn
	reCanInlineName
	reCanInlineCost
)

var reCannotInline = regexp.MustCompile(`^(.+?):(\d+):(\d+): cannot inline (\w+): (.*)`)

const (
	reCannotInlineFile = iota + 1
	reCannotInlineLine
	reCannotInlineColumn
	reCannotInlineName
	reCannotInlineReason
)

var reEscapesToHeap = regexp.MustCompile(`^(.+?):(\d+):(\d+): (\w+) escapes to heap:`)

const (
	reEscapesToHeapFile = iota + 1
	reEscapesToHeapLine
	reEscapesToHeapColumn
	reEscapesToHeapName
)
