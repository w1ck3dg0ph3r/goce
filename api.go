package main

import (
	"fmt"
	"io"

	"github.com/gofiber/fiber/v2"
	"mvdan.cc/gofumpt/format"

	"github.com/w1ck3dg0ph3r/goce/compilers"
	"github.com/w1ck3dg0ph3r/goce/parsers"
)

type API struct{}

func (api *API) GetCompilers(ctx *fiber.Ctx) error {
	type CompilersResponse []compilers.CompilerInfo
	list := compilers.List()
	return ctx.JSON(list)
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
		LangVersion: "1.20",
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

	compiler := compilers.Default()
	if req.Name != "" {
		compiler = compilers.Get(req.Name)
	}
	if compiler == nil {
		return fiber.ErrNotFound
	}

	compRes, err := compiler.Compile([]byte(req.Code))
	if err != nil {
		output, _ := io.ReadAll(compRes.BuildOutput)
		return ctx.JSON(Response{
			Errors: fmt.Sprintf("%s\n%s", output, err.Error()),
		})
	}

	parser := parsers.FindMatching(compRes)
	parseRes := parser.Parse(compRes)

	return ctx.JSON(Response{
		Result: &parseRes,
	})
}
