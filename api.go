package main

import (
	"fmt"
	"io"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/gofiber/fiber/v2"
	"mvdan.cc/gofumpt/format"

	"github.com/w1ck3dg0ph3r/goce/compilers"
	"github.com/w1ck3dg0ph3r/goce/parsers"
)

type API struct {
	CompilationCache *CompilationCache
}

const (
	CompilationCacheTTL = 2 * time.Hour
)

func (api *API) GetCompilers(ctx *fiber.Ctx) error {
	type CompilersResponse []compilers.CompilerInfo
	list := compilers.List()
	return ctx.JSON(CompilersResponse(list))
}

func (api *API) Format(ctx *fiber.Ctx) error {
	type Request struct {
		Name string `json:"name"`
		Code string `json:"code"`
	}
	type Response struct {
		Code   string `json:"code"`
		Errors string `json:"errors,omitempty"`
	}

	var req Request
	if err := ctx.BodyParser(&req); err != nil {
		fmt.Println(string(ctx.Body()))
		return err
	}

	formattedCode, err := format.Source([]byte(req.Code), format.Options{
		LangVersion: gofumptVersionForCompiler(req.Name),
		ExtraRules:  true,
	})
	res := Response{
		Code: string(formattedCode),
	}
	if err != nil {
		res.Errors = err.Error()
	}

	return ctx.JSON(res)
}

func (api *API) Compile(ctx *fiber.Ctx) error {
	type Request struct {
		Name string `json:"name"`
		Code string `json:"code"`
	}
	type Response struct {
		*parsers.Result
		Errors string `json:"errors,omitempty"`
	}

	var req Request
	if err := ctx.BodyParser(&req); err != nil {
		return err
	}
	code := []byte(req.Code)

	var cacheValue CompilationCacheValue
	if found, err := api.CompilationCache.Get(CompilationCacheKey{CompilerName: req.Name, Code: code}, &cacheValue); found {
		if cacheValue.Errors != "" {
			return ctx.JSON(Response{
				Errors: cacheValue.Errors,
			})
		}
		return ctx.JSON(Response{
			Result: &cacheValue.Result,
		})
	} else if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	compiler := compilers.Default()
	if req.Name != "" {
		compiler = compilers.Get(req.Name)
	}
	if compiler == nil {
		return fiber.NewError(fiber.StatusNotFound, "compiler not found: ", req.Name)
	}

	compRes, err := compiler.Compile(code)
	if err != nil {
		output, _ := io.ReadAll(compRes.BuildOutput)
		errors := fmt.Sprintf("%s\n%s", output, err.Error())
		if err := api.CompilationCache.Set(
			CompilationCacheKey{CompilerName: compRes.CompilerInfo.Name, Code: code},
			CompilationCacheValue{Errors: errors},
			CompilationCacheTTL,
		); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		return ctx.JSON(Response{
			Errors: errors,
		})
	}

	parser := parsers.FindMatching(compRes)
	if parser == nil {
		return fiber.NewError(fiber.StatusNotFound, "parser not found for go version: ", compRes.CompilerInfo.Version)
	}
	parseRes := parser.Parse(compRes)

	if err := api.CompilationCache.Set(
		CompilationCacheKey{CompilerName: compRes.CompilerInfo.Name, Code: code},
		CompilationCacheValue{Result: parseRes},
		CompilationCacheTTL,
	); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(Response{
		Result: &parseRes,
	})
}

func gofumptVersionForCompiler(name string) string {
	const defaultVeriosn = "1.20"

	if name == "" {
		return defaultVeriosn
	}
	compiler := compilers.Get(name)
	if compiler == nil {
		return defaultVeriosn
	}
	info, err := compiler.Info()
	if err != nil {
		return defaultVeriosn
	}
	ver, err := semver.NewVersion(info.Version)
	if err != nil {
		return defaultVeriosn
	}
	return fmt.Sprintf("%d.%d", ver.Major(), ver.Minor())
}
