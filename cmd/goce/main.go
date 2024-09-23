package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"github.com/w1ck3dg0ph3r/goce/api"
	"github.com/w1ck3dg0ph3r/goce/compilers"
	"github.com/w1ck3dg0ph3r/goce/config"
	"github.com/w1ck3dg0ph3r/goce/pkg/cache"
	"github.com/w1ck3dg0ph3r/goce/store"
	"github.com/w1ck3dg0ph3r/goce/ui"
)

var version string

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	app := fiber.New(fiber.Config{
		AppName:       "GoCE " + version,
		CaseSensitive: true,
		StrictRouting: true,
		ReadTimeout:   3 * time.Second,
		WriteTimeout:  3 * time.Second,
		IdleTimeout:   30 * time.Second,
	})

	app.Use(cors.New())
	app.Use(logger.New())
	app.Use(compress.New())
	app.Use(etag.New(etag.Config{Weak: true}))
	app.Use(sanityCheck())

	compilersSvc, err := compilers.New(&compilers.Config{
		SearchGoPath:            cfg.Compilers.SearchGoPath,
		SearchSDKPath:           cfg.Compilers.SearchSDKPath,
		LocalCompilers:          cfg.Compilers.LocalCompilers,
		AdditionalArchitectures: cfg.Compilers.AdditionalArchitectures,
		EnableModules:           cfg.Compilers.EnableModules,
	})
	if err != nil {
		fmt.Printf("compilers service: %v\n", err)
		os.Exit(1)
	}

	if err := os.Mkdir("data", os.ModeDir|os.ModePerm); err != nil && !errors.Is(err, os.ErrExist) {
		fmt.Printf("can't create data directory: %v", err.Error())
	}

	var compilationCache *cache.Cache[store.CompilationCacheKey, store.CompilationCacheValue]
	if cfg.Cache.Enabled {
		compilationCache, err = store.NewCompilationCache("data/cache.db")
		if err != nil {
			fmt.Printf("compilation cache: %v", err.Error())
			os.Exit(1)
		}
	}

	sharedCodeStore, err := store.NewSharedCode("data/shared.db")
	if err != nil {
		fmt.Printf("shared code store: %v", err.Error())
		os.Exit(1)
	}

	api := &api.API{
		Config: cfg,

		Compilers:        compilersSvc,
		CompilationCache: compilationCache,
		SharedCodeStore:  sharedCodeStore,
	}

	app.Get("/api/compilers", api.GetCompilers)
	app.Post("/api/format", api.Format)
	app.Post("/api/compile", api.Compile)
	app.Post("/api/shared", api.ShareCode)
	app.Get("/api/shared/:id", api.GetSharedCode)

	app.Use("/", serveUI())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	doneCh := make(chan struct{})
	go func() {
		<-sigCh
		fmt.Printf("shutdown signal received, terminating...\n")
		if err := app.Shutdown(); err != nil {
			fmt.Printf("shutdown: %v\n", err)
		}
		if compilationCache != nil {
			compilationCache.Close()
		}
		if sharedCodeStore != nil {
			sharedCodeStore.Close()
		}
		doneCh <- struct{}{}
	}()
	if err := app.Listen(cfg.Listen); err != nil {
		fmt.Printf("error: %v\n", err)
	}
	<-doneCh
}

func serveUI() fiber.Handler {
	return filesystem.New(filesystem.Config{
		Root:         http.FS(ui.DistFS),
		PathPrefix:   "dist",
		Index:        "index.html",
		NotFoundFile: "dist/index.html",
		MaxAge:       60,
	})
}

func sanityCheck() fiber.Handler {
	const maxContentLength = 64 << 10
	errInsane := fiber.NewError(fiber.StatusBadRequest, "request too long")

	return func(ctx *fiber.Ctx) error {
		if ctx.Request().Header.ContentLength() > maxContentLength {
			return errInsane
		}

		if len(ctx.Request().Body()) > maxContentLength {
			return errInsane
		}

		return ctx.Next()
	}
}
