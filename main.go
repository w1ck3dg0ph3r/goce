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
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

//go:embed ui/dist
var distFS embed.FS

func main() {
	app := fiber.New(fiber.Config{
		AppName:       "GOCE v0.0.1",
		CaseSensitive: true,
		StrictRouting: true,
		ReadTimeout:   3 * time.Second,
		WriteTimeout:  3 * time.Second,
		IdleTimeout:   15 * time.Second,
	})

	app.Use(cors.New())
	app.Use(logger.New())
	app.Use(compress.New())

	api := &API{}

	app.Get("/api/compilers", api.GetCompilers)
	app.Post("/api/format", api.Format)
	app.Post("/api/compile", api.Compile)

	app.Use("/", serveUI())

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Printf("shutdown signal received, terminating...\n")
		if err := app.Shutdown(); err != nil {
			fmt.Printf("shutdown: %v\n", err)
		}
	}()
	if err := app.Listen(":9000"); err != nil {
		fmt.Printf("error: %v\n", err)
	}
}

func serveUI() fiber.Handler {
	return filesystem.New(filesystem.Config{
		Root:       http.FS(distFS),
		PathPrefix: "ui/dist",
		Index:      "/index.html",
		MaxAge:     60,
	})
}
