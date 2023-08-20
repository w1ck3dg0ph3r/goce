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

func (c *localCompiler) Compile(ctx context.Context, config CompilerConfig, code []byte) (Result, error) {
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

	buildEnv := os.Environ()
	if config.Platform != c.info.Platform {
		buildEnv = append(buildEnv, fmt.Sprintf("GOOS=%s", config.Platform))
	}
	if config.Architecture != c.info.Architecture {
		buildEnv = append(buildEnv, fmt.Sprintf("GOARCH=%s", config.Architecture))
	}

	res := Result{
		CompilerInfo:   c.info,
		SourceFilename: mainFilename,
		SourceCode:     code,
	}

	cmd := exec.CommandContext(ctx, c.GoPath, "build", "-o", os.DevNull, goFilename)
	cmd.Dir = buildDir
	cmd.Env = buildEnv
	output, err := cmd.CombinedOutput()
	output, _ = bytes.CutPrefix(output, []byte("# command-line-arguments\n"))
	res.BuildOutput = output
	if err != nil {
		return res, fmt.Errorf("build: %w", err)
	}

	cmd = exec.CommandContext(ctx, c.GoPath, "build", "-o", "main", "-trimpath", "-gcflags", "-m=2", goFilename)
	cmd.Dir = buildDir
	cmd.Env = buildEnv
	output, err = cmd.CombinedOutput()
	output, _ = bytes.CutPrefix(output, []byte("# command-line-arguments\n"))
	res.BuildOutput = output
	if err != nil {
		return res, fmt.Errorf("debug: %w", err)
	}

	cmd = exec.CommandContext(ctx, c.GoPath, "tool", "objdump", "-s", "^main\\.", "main")
	cmd.Dir = buildDir
	output, err = cmd.CombinedOutput()
	if err != nil {
		return res, fmt.Errorf("objdump: %w", err)
	}
	res.ObjdumpOutput = output

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
	match := reGoVersion.FindSubmatch(out)
	if match == nil {
		return CompilerInfo{}, fmt.Errorf("cant parse go version: %q", string(out))
	}
	c.info = CompilerInfo{
		Version:      string(match[reGoVersion_Version]),
		Platform:     string(match[reGoVersion_Platform]),
		Architecture: string(match[reGoVersion_Architecture]),
	}
	return c.info, nil
}

var reGoVersion = regexp.MustCompile(`go(\d+\.\d+(\.\d+)?)\s+(\w+)/(\w+)`)

const (
	reGoVersion_Version = iota + 1
	reGoVersion_VersionPatch
	reGoVersion_Platform
	reGoVersion_Architecture
)
