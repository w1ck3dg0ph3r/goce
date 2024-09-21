package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/w1ck3dg0ph3r/goce/compilers"
	"github.com/w1ck3dg0ph3r/goce/parsers"
)

func TestGoce(t *testing.T) {
	binary := buildGoce(t)
	goce := startGoce(t, binary)
	defer stopGoce(t, goce)

	var availableCompilers []struct {
		Name         string `json:"name"`
		Version      string `json:"version"`
		Platform     string `json:"platform"`
		Architecture string `json:"architecture"`
	}

	t.Run("ListCompilers", func(t *testing.T) {
		status, err := request("GET", "/api/compilers", nil, &availableCompilers)
		if err != nil {
			t.Error(err)
		}
		if status != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, status)
		}
		fmt.Printf("available compilers:\n")
		for _, c := range availableCompilers {
			fmt.Printf("- %s\n", c.Name)
		}
	})

	if len(availableCompilers) == 0 {
		t.Fatal("no available compilers")
	}

	t.Run("FormatCode", func(t *testing.T) {
		req := struct {
			Name string `json:"name"`
			Code string `json:"code"`
		}{
			Name: availableCompilers[0].Name,
			Code: readTestFile("format_in.go.txt"),
		}
		var res struct {
			Code   string `json:"code"`
			Errors string `json:"errors,omitempty"`
		}
		status, err := request("POST", "/api/format", req, &res)
		if err != nil {
			t.Error(err)
		}
		if status != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, status)
		}
		if res.Errors != "" {
			t.Errorf("unexpected format errors: %q", res.Errors)
		}
		if res.Code != readTestFile("format_out.go.txt") {
			t.Errorf("unexpected format result: %q", res.Code)
		}
	})

	t.Run("Compile", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			req := struct {
				Name    string                    `json:"name"`
				Options compilers.CompilerOptions `json:"options"`
				Code    string                    `json:"code"`
			}{
				Name: availableCompilers[0].Name,
				Code: readTestFile("example.go"),
			}
			var res struct {
				BuildFailed bool   `json:"buildFailed"`
				BuildOutput string `json:"buildOutput"`
				parsers.Result
			}
			status, err := request("POST", "/api/compile", req, &res)
			if err != nil {
				t.Error(err)
			}
			if status != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, status)
			}
			if res.BuildFailed {
				t.Errorf("expected build to succeed")
			}
			if res.BuildOutput == "" {
				t.Errorf("expected build output")
			}
			if res.Assembly == "" {
				t.Errorf("expected assembly")
			}
		})

		t.Run("Failure", func(t *testing.T) {
			req := struct {
				Name    string                    `json:"name"`
				Options compilers.CompilerOptions `json:"options"`
				Code    string                    `json:"code"`
			}{
				Name: availableCompilers[0].Name,
				Code: readTestFile("invalid.go.txt"),
			}
			var res struct {
				BuildFailed bool   `json:"buildFailed"`
				BuildOutput string `json:"buildOutput"`
				parsers.Result
			}
			status, err := request("POST", "/api/compile", req, &res)
			if err != nil {
				t.Error(err)
			}
			if status != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, status)
			}
			if !res.BuildFailed {
				t.Errorf("expected build to fail")
			}
			if !strings.Contains(res.BuildOutput, `"math" imported and not used`) ||
				!strings.Contains(res.BuildOutput, `undefined: fmt`) {
				t.Errorf("expected valid error messages in build output, got %q", res.BuildOutput)
			}
		})
	})

	t.Run("Share", func(t *testing.T) {
		t.Run("NotFound", func(t *testing.T) {
			status, _ := request("GET", "/api/shared/3fH9yF8z", nil, nil)
			if status != http.StatusNotFound {
				t.Errorf("expected status %d, got %d", http.StatusNotFound, status)
			}
		})

		var sharedID string
		sharedCode := readTestFile("example.go")

		t.Run("ShareCode", func(t *testing.T) {
			var res struct {
				ID string `json:"id"`
			}
			status, err := request("POST", "/api/shared", sharedCode, &res)
			if status != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusNotFound, status)
			}
			if err != nil {
				t.Error(err)
			}
			if res.ID == "" {
				t.Errorf("expected shared id")
			}
			sharedID = res.ID
		})

		t.Run("GetShared", func(t *testing.T) {
			var res string
			status, err := request("GET", "/api/shared/"+sharedID, nil, &res)
			if status != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusNotFound, status)
			}
			if err != nil {
				t.Error(err)
			}
			if res != sharedCode {
				t.Errorf("expected to get shared code, got %q", res)
			}
		})
	})
}

func request(method, path string, req, res any) (int, error) {
	const base = "http://localhost:9000"
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	var body io.Reader
	if req != nil {
		switch req := req.(type) {
		case string:
			buf := []byte(req)
			body = bytes.NewReader(buf)
		default:
			buf, err := json.Marshal(req)
			if err != nil {
				return 0, fmt.Errorf("encode request: %w", err)
			}
			body = bytes.NewReader(buf)
		}
	}
	httpReq, err := http.NewRequestWithContext(ctx, method, base+path, body)
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}
	if body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	httpRes, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return 0, fmt.Errorf("do request: %w", err)
	}
	defer httpRes.Body.Close()
	if res != nil && httpRes.StatusCode/100 == 2 {
		switch res := res.(type) {
		case *string:
			buf, err := io.ReadAll(httpRes.Body)
			if err != nil {
				return httpRes.StatusCode, fmt.Errorf("read response: %w", err)
			}
			*res = string(buf)
		default:
			if err := json.NewDecoder(httpRes.Body).Decode(res); err != nil {
				return httpRes.StatusCode, fmt.Errorf("decode response: %w", err)
			}
		}
	}
	return httpRes.StatusCode, nil
}

func buildGoce(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err.Error())
	}
	pkg := filepath.Join(cwd, "..")
	binary := filepath.Join(t.TempDir(), "goce")
	fmt.Printf("building goce: %q\n", binary)
	cmd := exec.Command("go", "build", "-o", binary, pkg)
	cmd.Env = os.Environ()
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	return binary
}

func startGoce(t *testing.T, binary string) *exec.Cmd {
	t.Helper()
	cmd := exec.Command(binary)
	cmd.Dir = filepath.Dir(binary)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env,
		"GOCE_CACHE_ENABLED=false",
		"GOCE_COMPILERS_ADDITIONAL_ARCHITECTURES=false",
	)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	fmt.Printf("starting goce: %q\n", binary)
	if err := cmd.Start(); err != nil {
		t.Fatal(err.Error())
	}
	waitForAPI(t, "http://localhost:9000")
	return cmd
}

func stopGoce(t *testing.T, cmd *exec.Cmd) {
	t.Helper()
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatal(err.Error())
	}
	if err := cmd.Wait(); err != nil {
		t.Fatal(err.Error())
	}
}

func waitForAPI(t *testing.T, url string) {
	t.Helper()
	const timeout = 15 * time.Second
	fmt.Printf("waiting for goce api: %q\n", url)
	start := time.Now()
	for {
		time.Sleep(100 * time.Millisecond)
		req, err := http.NewRequest("GET", url+"/ok", nil)
		if err != nil {
			continue
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			continue
		}
		res.Body.Close()
		if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusNotFound {
			break
		}
		if time.Since(start) > timeout {
			t.Fatal("timed out waiting for api")
		}
	}
}

func readTestFile(fn string) string {
	path := filepath.Join("testdata", fn)
	b, err := os.ReadFile(path)
	if err != nil {
		panic(fmt.Errorf("read testdata file: %w", err))
	}
	return string(b)
}
