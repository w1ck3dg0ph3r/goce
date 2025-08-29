package main

import (
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/rs/zerolog"

	"github.com/w1ck3dg0ph3r/goce/api"
	"github.com/w1ck3dg0ph3r/goce/compilers"
	"github.com/w1ck3dg0ph3r/goce/config"
	"github.com/w1ck3dg0ph3r/goce/pkg/cache"
	"github.com/w1ck3dg0ph3r/goce/store"
	"github.com/w1ck3dg0ph3r/goce/ui"
)

var version string

func main() {
	log := zerolog.New(os.Stdout).Level(zerolog.InfoLevel)
	log.Info().Str("version", version).Msg("starting goce")

	cfg, err := config.Read()
	if err != nil {
		log.Error().Err(err).Msg("read config failed")
		os.Exit(1)
	}

	app := fiber.New(fiber.Config{
		AppName:               "GoCE " + version,
		DisableStartupMessage: true,
		CaseSensitive:         true,
		StrictRouting:         true,
		ReadTimeout:           3 * time.Second,
		WriteTimeout:          3 * time.Second,
		IdleTimeout:           30 * time.Second,
	})

	app.Use(fiberzerolog.New(fiberzerolog.Config{
		Logger: &log,
		Fields: []string{
			fiberzerolog.FieldMethod,
			fiberzerolog.FieldURL,
			fiberzerolog.FieldStatus,
			fiberzerolog.FieldLatency,
		},
		Messages: []string{
			"request server error",
			"request client error",
			"request success",
		},
	}))
	app.Use(cors.New())
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
		log.Error().Err(err).Msg("compilers service failed")
		os.Exit(1)
	}

	if err := os.Mkdir("data", os.ModeDir|os.ModePerm); err != nil && !errors.Is(err, os.ErrExist) {
		log.Warn().Err(err).Msg("can't create data directory")
	}

	var compilationCache *cache.Cache[store.CompilationCacheKey, store.CompilationCacheValue]
	if cfg.Cache.Enabled {
		compilationCache, err = store.NewCompilationCache("data/cache.db")
		if err != nil {
			log.Error().Err(err).Msg("compilation cache failed")
			os.Exit(1)
		}
	}

	sharedCodeStore, err := store.NewSharedCode("data/shared.db")
	if err != nil {
		log.Error().Err(err).Msg("shared code store failed")
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
		log.Info().Msg("shutdown signal received, terminating...\n")
		if err := app.Shutdown(); err != nil {
			log.Error().Err(err).Msg("shutdown failed")
		}
		if compilationCache != nil {
			compilationCache.Close()
		}
		if sharedCodeStore != nil {
			sharedCodeStore.Close()
		}
		doneCh <- struct{}{}
	}()

	log.Info().Str("address", cfg.Listen).Msg("listening")
	if err := app.Listen(cfg.Listen); err != nil {
		log.Error().Err(err).Msg("listen failed")
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
