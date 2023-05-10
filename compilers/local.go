package compilers

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

type localCompiler struct {
	GoPath string

	info CompilerInfo
}

func (c *localCompiler) Compile(ctx context.Context, code []byte) (Result, error) {
	tmpDir := filepath.Join(os.TempDir(), "goce")
	if err := os.MkdirAll(tmpDir, 0o777); err != nil {
		return Result{}, fmt.Errorf("create tmp dir: %w", err)
	}
	buildDir, err := os.MkdirTemp(tmpDir, "build-")
	if err != nil {
		return Result{}, fmt.Errorf("create tmp dir: %w", err)
	}
	defer os.RemoveAll(buildDir)

	const goFilename = "main.go"
	mainFilename := filepath.Join(buildDir, goFilename)
	fmain, err := os.Create(mainFilename)
	if err != nil {
		return Result{}, fmt.Errorf("create source file: %w", err)
	}
	if _, err := fmain.Write(code); err != nil {
		return Result{}, fmt.Errorf("write source file: %w", err)
	}
	fmain.Close()

	buildOutput := &bytes.Buffer{}
	objdumpOutput := &bytes.Buffer{}

	cmd := exec.CommandContext(ctx, c.GoPath, "build", "-o", "main", "-trimpath", "-gcflags", "-m=2", goFilename)
	cmd.Dir = buildDir
	cmd.Stdout = buildOutput
	cmd.Stderr = buildOutput

	res := Result{
		CompilerInfo:   c.info,
		SourceFilename: mainFilename,
		SourceCode:     code,
		BuildOutput:    buildOutput,
		ObjdumpOutput:  objdumpOutput,
	}

	if err := cmd.Run(); err != nil {
		return res, fmt.Errorf("build: %w", err)
	}

	cmd = exec.CommandContext(ctx, c.GoPath, "tool", "objdump", "-s", "^main\\.", "main")
	cmd.Dir = buildDir
	cmd.Stdout = objdumpOutput
	cmd.Stderr = objdumpOutput

	if err := cmd.Run(); err != nil {
		return res, fmt.Errorf("objdump: %w", err)
	}

	return res, nil
}

func (c *localCompiler) Info() (CompilerInfo, error) {
	if c.info.Version != "" {
		return c.info, nil
	}

	cmd := exec.Command(c.GoPath, "version")
	out, err := cmd.Output()
	if err != nil {
		return CompilerInfo{}, fmt.Errorf("cant run go version: %w", err)
	}
	out = bytes.TrimPrefix(out, []byte("go version "))
	out = bytes.TrimSpace(out)
	name := string(out)
	match := reGoVersion.FindSubmatch(out)
	if match == nil {
		return CompilerInfo{}, fmt.Errorf("cant parse go version: %q", string(out))
	}
	c.info = CompilerInfo{
		Name:     name,
		Version:  string(match[reGoVersionVersion]),
		Platform: string(match[reGoVersionPlatform]),
	}
	return c.info, nil
}

var reGoVersion = regexp.MustCompile(`go(\d+\.\d+(\.\d+)?)\s+([\w\/]+)`)

const (
	reGoVersionVersion = iota + 1
	reGoVersionVersionPatch
	reGoVersionPlatform
)
