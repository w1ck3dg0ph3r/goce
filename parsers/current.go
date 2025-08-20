package parsers

import (
	"bufio"
	"bytes"
	"io"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/w1ck3dg0ph3r/goce/compilers"
)

type currentParser struct{}

func (currentParser) Parse(output compilers.Result) Result {
	var res Result
	parseBuildOutput(&res, output.SourceCode, bytes.NewReader(output.BuildOutput))
	parseObjdumpOutput(&res, bytes.NewReader(output.ObjdumpOutput))
	return res
}

func parseBuildOutput(res *Result, sourceCode []byte, output io.Reader) {
	sc := bufio.NewScanner(output)

	mainFilenameBytes := []byte("./main.go")
	sourceLines := bytes.Split(sourceCode, []byte{'\n'})

	for sc.Scan() {
		if err := sc.Err(); err != nil {
			break
		}

		line := sc.Bytes()
		if len(line) == 0 || isComment(line) {
			continue
		}

		var match [][]byte

		if match = reBuildLine.FindSubmatch(line); match == nil {
			continue
		}
		fileName := match[reBuildLine_FileName]
		location := Location{
			Line:   mustParseInt(match[reBuildLine_Line]),
			Column: mustParseInt(match[reBuildLine_Column]),
		}
		text := match[reBuildLine_Text]
		if !bytes.Equal(fileName, mainFilenameBytes) {
			continue
		}
		if indentLevel(text) > 0 {
			continue
		}

		// Can Inline
		if match = reCanInline.FindSubmatch(text); match != nil {
			name := string(match[reCanInline_Name])
			fc := InliningAnalysis{
				Diagnostic: Diagnostic{
					Type:  DiagnosticInliningAnalysis,
					Range: makeRange(locationToUnicode(sourceLines, location), len(name)),
				},
				Name:      name,
				CanInline: true,
			}
			cost, _ := strconv.Atoi(string(match[reCanInline_Cost]))
			fc.Cost = cost
			res.Diagnostics = append(res.Diagnostics, fc)
		}

		// Cannot Inline
		if match = reCannotInline.FindSubmatch(text); match != nil {
			name := string(match[reCannotInline_Name])
			fc := InliningAnalysis{
				Diagnostic: Diagnostic{
					Type:  DiagnosticInliningAnalysis,
					Range: makeRange(locationToUnicode(sourceLines, location), len(name)),
				},
				Name:      string(match[reCannotInline_Name]),
				CanInline: false,
				Reason:    string(match[reCannotInline_Reason]),
			}
			res.Diagnostics = append(res.Diagnostics, fc)
		}

		// Inlining Call
		if match = reInliningCall.FindSubmatch(text); match != nil {
			col := location.Column
			line := sourceLines[location.Line-1]
			name := match[reInliningCall_Name]
			nameLen := len(name)
			if bytes.HasSuffix(line[:col-1], name) {
				col -= nameLen
			} else {
				nameLen = suffixWordLength(line[:col-1])
				col -= nameLen
			}
			ic := InlinedCall{
				Diagnostic: Diagnostic{
					Type: DiagnosticInlinedCall,
					Range: makeRange(locationToUnicode(
						sourceLines,
						Location{
							Line:   location.Line,
							Column: col,
						},
					), nameLen),
				},
				Name: string(name),
			}
			res.Diagnostics = append(res.Diagnostics, ic)
		}

		// Heap escapes
		if match = reEscapesToHeap.FindSubmatch(text); match != nil {
			line := sourceLines[location.Line-1]
			name := match[reEscapesToHeap_Name]
			if !bytes.HasSuffix(text, []byte(":")) {
				he := HeapEscape{
					Diagnostic: Diagnostic{
						Type:  DiagnosticHeapEscape,
						Range: makeRange(locationToUnicode(sourceLines, location), 0),
					},
				}
				if bytes.HasPrefix(line[location.Column-1:], name) {
					he.Name = string(match[reEscapesToHeap_Name])
					he.Range.End.Column += len(he.Name)
				} else {
					he.Message = string(text)
				}
				res.Diagnostics = append(res.Diagnostics, he)
			}

			// Go versions prior to 1.20 seem to report column-1 for heap escapes
			if bytes.HasPrefix(line[location.Column:], name) {
				location.Column += 1
				he := HeapEscape{
					Diagnostic: Diagnostic{
						Type:  DiagnosticHeapEscape,
						Range: makeRange(locationToUnicode(sourceLines, location), 0),
					},
					Name: string(match[reEscapesToHeap_Name]),
				}
				res.Diagnostics = append(res.Diagnostics, he)
			}
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

		line := sc.Bytes()
		var match [][]byte

		if bytes.HasPrefix(line, []byte("TEXT")) {
			assembly.Write([]byte(sc.Text()))
			assembly.WriteRune('\n')
			assemblyLine++
		}

		if match = reAssembly.FindSubmatch(line); match != nil {
			assembly.Write(match[reAssembly_Address])
			assembly.WriteRune('\t')
			assembly.Write(match[reAssembly_Code])
			assembly.WriteRune('\n')
			assemblyLine++
			if bytes.Equal(match[reAssembly_File], mainFilenameBytes) {
				lineNumber, _ := strconv.Atoi(string(match[reAssembly_Line]))
				if lineNumber != lastSourceLine {
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

func isComment(line []byte) bool {
	return len(line) > 0 && line[0] == '#'
}

func indentLevel(line []byte) int {
	level := 0
	for _, c := range line {
		if c != ' ' {
			break
		}
		level++
	}
	return level / 2
}

func suffixWordLength(s []byte) int {
	blen := len(s)
	for p := blen - 1; p >= 0; p-- {
		if !unicode.IsLetter(rune(s[p])) {
			return blen - 1 - p
		}
	}
	return blen
}

func makeRange(start Location, length int) Range {
	end := start
	end.Column += length
	return Range{Start: start, End: end}
}

// locationToUnicode tranforms loc's Column to runes instead of bytes.
func locationToUnicode(sourceLines [][]byte, loc Location) Location {
	line := sourceLines[loc.Line-1]
	if loc.Column > len(line) {
		return loc
	}
	line = line[:loc.Column]
	column := 0
	for len(line) > 0 {
		r, size := utf8.DecodeLastRune(line)
		if r == utf8.RuneError {
			break
		}
		line = line[:len(line)-size]
		column++
	}
	return Location{Line: loc.Line, Column: column}
}

func mustParseInt(s []byte) int {
	i, err := strconv.Atoi(string(s))
	if err != nil {
		panic(err)
	}
	return i
}

var reAssembly = regexp.MustCompile(`^\s+(.+?):(\d+)\t+(\w+)\t+([^\t]+)\t+(.*?)\t*$`)

const (
	reAssembly_File = iota + 1
	reAssembly_Line
	reAssembly_Address
	reAssembly_Opcodes
	reAssembly_Code
)

var reBuildLine = regexp.MustCompile(`^(.+?):(\d+):(\d+): (.*)`)

const (
	reBuildLine_FileName = iota + 1
	reBuildLine_Line
	reBuildLine_Column
	reBuildLine_Text
)

var reCanInline = regexp.MustCompile(`^can inline (\w+) with cost (\d+)`)

const (
	reCanInline_Name = iota + 1
	reCanInline_Cost
)

var reCannotInline = regexp.MustCompile(`^cannot inline (\w+): (.*)`)

const (
	reCannotInline_Name = iota + 1
	reCannotInline_Reason
)

var reInliningCall = regexp.MustCompile(`^inlining call to (.*)`)

const (
	reInliningCall_Name = iota + 1
)

var reEscapesToHeap = regexp.MustCompile(`^([\w_&{}]+) escapes to heap$`)

const (
	reEscapesToHeap_Name = iota + 1
)
