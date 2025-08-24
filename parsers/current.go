package parsers

import (
	"bufio"
	"bytes"
	"encoding/json"
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
	parseJSON(&res, output.BuildJSON)
	return res
}

func parseBuildOutput(res *Result, sourceCode []byte, output io.Reader) {
	sc := bufio.NewScanner(output)

	mainFilenameBytes := []byte("./main.go")
	sourceLines := bytes.Split(sourceCode, []byte{'\n'})

	buildOutput := &strings.Builder{}
	assembly := strings.Builder{}
	lastSourceLine := 0
	assemblyLine := 0

	for sc.Scan() {
		if err := sc.Err(); err != nil {
			break
		}

		line := sc.Bytes()
		if len(line) == 0 || isComment(line) {
			continue
		}

		var match [][]byte

		if bytes.Contains(line, []byte(" STEXT ")) {
			continue
		}

		if match = reAssembly.FindSubmatch(line); match != nil {
			assembly.Write(match[reAssembly_Address])
			assembly.WriteRune('\t')
			bytesReplace(match[reAssembly_Code], '\t', ' ')
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
			continue
		}

		if match = reBuildLine.FindSubmatch(line); match != nil {
			buildOutput.Write(line)
			buildOutput.WriteByte('\n')
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

				he := HeapEscape{
					Diagnostic: Diagnostic{
						Type:  DiagnosticHeapEscape,
						Range: makeRange(locationToUnicode(sourceLines, location), 1),
					},
				}
				if bytes.HasPrefix(line[location.Column-1:], name) {
					he.Name = string(match[reEscapesToHeap_Name])
					he.Range.End.Column = he.Range.Start.Column + len(he.Name)
				} else if pos := bytes.Index(line, match[reEscapesToHeap_Name]); pos != -1 {
					he.Name = string(match[reEscapesToHeap_Name])
					he.Range.Start.Column = pos + 1
					he.Range.End.Column = he.Range.Start.Column + len(he.Name)
				} else {
					he.Message = string(text)
				}
				res.Diagnostics = append(res.Diagnostics, he)

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

	res.Assembly = assembly.String()
	res.BuildOutput = buildOutput.String()
}

type bjsonHeader struct {
	File    string `json:"file"`
	Version int    `json:"version"`
}

type bjsonDiagnostic struct {
	Code    string     `json:"code"`
	Message string     `json:"message"`
	Range   bjsonRange `json:"range"`
}

type bjsonRange struct {
	Start bjsonPosition `json:"start"`
	End   bjsonPosition `json:"end"`
}

type bjsonPosition struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

func parseJSON(res *Result, data []byte) {
	dec := json.NewDecoder(bytes.NewReader(data))
	var header bjsonHeader
	if err := dec.Decode(&header); err != nil {
		return
	}
	for dec.More() {
		var d bjsonDiagnostic
		if err := dec.Decode(&d); err != nil {
			continue
		}
		switch d.Code {
		case "isInBounds", "isSliceInBounds":
			diag := Diagnostic{
				Type: DiagnosticBoundsCheck,
				Range: Range{
					Start: Location{Line: d.Range.Start.Line, Column: d.Range.Start.Character},
					End:   Location{Line: d.Range.End.Line, Column: d.Range.End.Character},
				},
			}
			if diag.Range.End.Column == diag.Range.Start.Column {
				diag.Range.End.Column += 1
			}
			res.Diagnostics = append(res.Diagnostics, diag)

		// Other known codes:
		case "canInlineFunction":
		case "cannotInlineCall":
		case "cannotInlineFunction":
		case "escape":
		case "escapes":
		case "leak":
		case "nilcheck":
		}
	}
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

func bytesReplace(bs []byte, old, new byte) {
	for i, b := range bs {
		if b == old {
			bs[i] = new
		}
	}
}

var reAssembly = regexp.MustCompile(`^\t(\w+) \d+ \(([^:]+):(\d+)\)\t(.*)`)

const (
	reAssembly_Address = iota + 1
	reAssembly_File
	reAssembly_Line
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

var reEscapesToHeap = regexp.MustCompile(`^(.+) escapes to heap:$`)

const (
	reEscapesToHeap_Name = iota + 1
)
