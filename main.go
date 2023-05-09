package main

import (
	"embed"
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
)

//go:embed ui/dist
var distFS embed.FS

func main() {
	app := fiber.New(fiber.Config{
		AppName:       "GoCE v0.0.1",
		CaseSensitive: true,
		StrictRouting: true,
		ReadTimeout:   3 * time.Second,
		WriteTimeout:  3 * time.Second,
		IdleTimeout:   15 * time.Second,
	})

	app.Use(cors.New())
	app.Use(logger.New())
	app.Use(compress.New())
	app.Use(etag.New(etag.Config{Weak: true}))

	compilationCache, err := NewCompilationCache("cache.db")
	if err != nil {
		fmt.Printf("compilation cache: %v", err.Error())
		os.Exit(1)
	}

	sharedCodeStore, err := NewSharedStore("shared.db")
	if err != nil {
		fmt.Printf("shared code store: %v", err.Error())
		os.Exit(1)
	}

	api := &API{
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
		compilationCache.Close()
		doneCh <- struct{}{}
	}()
	if err := app.Listen(":9000"); err != nil {
		fmt.Printf("error: %v\n", err)
	}
	<-doneCh
}

func serveUI() fiber.Handler {
	return filesystem.New(filesystem.Config{
		Root:         http.FS(distFS),
		PathPrefix:   "ui/dist",
		Index:        "index.html",
		NotFoundFile: "ui/dist/index.html",
		MaxAge:       60,
	})
}
