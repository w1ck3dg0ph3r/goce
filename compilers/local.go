package compilers

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type localCompiler struct {
	GoPath        string
	EnableModules bool

	info CompilerInfo
}

func (c *localCompiler) Compile(ctx context.Context, config CompilerConfig, code []byte) (Result, error) {
	info, err := c.Info()
	if err != nil {
		return Result{}, fmt.Errorf("get compiler info: %w", err)
	}
	run := &localRun{
		GoPath: c.GoPath,
		Code:   code,
		Info:   info,
		Config: config,
	}
	if err := run.Prepare(); err != nil {
		return Result{}, fmt.Errorf("prepare: %w", err)
	}
	defer run.Close()

	if c.EnableModules {
		if err := run.InitModules(ctx); err != nil {
			return Result{}, fmt.Errorf("init modules: %w", err)
		}
	}

	res := Result{
		CompilerInfo:   info,
		SourceFilename: run.mainBasename,
		SourceCode:     code,
	}
	if res.BuildOutput, err = run.Build(ctx); err != nil {
		return res, fmt.Errorf("build: %w", err)
	}
	if res.ObjdumpOutput, err = run.Objdump(ctx); err != nil {
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
		return CompilerInfo{}, fmt.Errorf("%w: go version: %w", ErrInvalidPath, err)
	}
	out = bytes.TrimPrefix(out, []byte("go version "))
	out = bytes.TrimSpace(out)
	match := reGoVersion.FindSubmatch(out)
	if match == nil {
		return CompilerInfo{}, fmt.Errorf("%w: go version: %q", ErrInvalidPath, string(out))
	}
	c.info = CompilerInfo{
		Version:      string(match[reGoVersion_Version]),
		Platform:     string(match[reGoVersion_Platform]),
		Architecture: string(match[reGoVersion_Architecture]),
	}
	return c.info, nil
}

type localRun struct {
	GoPath string
	Code   []byte
	Info   CompilerInfo
	Config CompilerConfig

	buildDir     string
	buildEnv     []string
	mainBasename string
}

func (r *localRun) Prepare() error {
	tmpDir := filepath.Join(os.TempDir(), "goce")
	if err := os.MkdirAll(tmpDir, 0o777); err != nil {
		return fmt.Errorf("create tmp dir: %w", err)
	}
	buildDir, err := os.MkdirTemp(tmpDir, "build-")
	if err != nil {
		return fmt.Errorf("create tmp dir: %w", err)
	}
	r.buildDir = buildDir

	r.mainBasename = "main.go"
	mainFilename := filepath.Join(buildDir, r.mainBasename)
	fmain, err := os.Create(mainFilename)
	if err != nil {
		return fmt.Errorf("create source file: %w", err)
	}
	if _, err := fmain.Write(r.Code); err != nil {
		return fmt.Errorf("write source file: %w", err)
	}
	fmain.Close()

	return nil
}

func (r *localRun) Close() {
	os.RemoveAll(r.buildDir)
}

func (r *localRun) InitModules(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, r.GoPath, "mod", "init", "goce-build")
	cmd.Dir = r.buildDir
	cmd.Env = r.BuildEnv()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go mod init: %w", err)
	}
	cmd = exec.CommandContext(ctx, r.GoPath, "mod", "tidy")
	cmd.Dir = r.buildDir
	cmd.Env = r.BuildEnv()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go mod tidy: %w", err)
	}
	return nil
}

func (r *localRun) Build(ctx context.Context) ([]byte, error) {
	if errors := r.buildErrors(ctx); errors != nil {
		return errors, ErrBuildFailed
	}

	args := []string{"build", "-o", "main", "-trimpath", "-gcflags"}
	var gcflags []string
	if r.Config.Options.DisableInlining {
		gcflags = append(gcflags, "-l")
	}
	if r.Config.Options.DisableOptimizations {
		gcflags = append(gcflags, "-N")
	}
	gcflags = append(gcflags, "-m=2")
	args = append(args, strings.Join(gcflags, " "))
	args = append(args, r.mainBasename)
	cmd := exec.CommandContext(ctx, r.GoPath, args...)
	cmd.Dir = r.buildDir
	cmd.Env = r.BuildEnv()
	output, err := cmd.CombinedOutput()
	output, _ = bytes.CutPrefix(output, []byte("# command-line-arguments\n"))
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (r *localRun) Objdump(ctx context.Context) ([]byte, error) {
	cmd := exec.CommandContext(ctx, r.GoPath, "tool", "objdump", "-s", "^main\\.", "main")
	cmd.Dir = r.buildDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (r *localRun) buildErrors(ctx context.Context) []byte {
	cmd := exec.CommandContext(ctx, r.GoPath, "build", "-o", os.DevNull, r.mainBasename)
	cmd.Dir = r.buildDir
	cmd.Env = r.BuildEnv()
	output, err := cmd.CombinedOutput()
	if err != nil {
		output, _ = bytes.CutPrefix(output, []byte("# command-line-arguments\n"))
		return output
	}
	return nil
}

func (r *localRun) BuildEnv() []string {
	if r.buildEnv != nil {
		return r.buildEnv
	}
	e := os.Environ()
	if r.Config.Platform != r.Info.Platform {
		e = append(e, fmt.Sprintf("GOOS=%s", r.Config.Platform))
	}
	if r.Config.Architecture != r.Info.Architecture {
		e = append(e, fmt.Sprintf("GOARCH=%s", r.Config.Architecture))
	}
	if r.Config.Options.ArchitectureLevel != "" {
		switch r.Config.Architecture {
		case "amd64":
			e = append(e, fmt.Sprintf("GOAMD64=%s", r.Config.Options.ArchitectureLevel))
		case "ppc64":
			e = append(e, fmt.Sprintf("GOPPC64=%s", r.Config.Options.ArchitectureLevel))
		case "386":
			e = append(e, fmt.Sprintf("GO386=%s", r.Config.Options.ArchitectureLevel))
		case "arm":
			e = append(e, fmt.Sprintf("GOARM=%s", r.Config.Options.ArchitectureLevel))
		}
	}
	r.buildEnv = e
	return r.buildEnv
}

var reGoVersion = regexp.MustCompile(`go(\d+\.\d+(\.\d+)?)\s+(\w+)/(\w+)`)

const (
	reGoVersion_Version = iota + 1
	reGoVersion_VersionPatch
	reGoVersion_Platform
	reGoVersion_Architecture
)
