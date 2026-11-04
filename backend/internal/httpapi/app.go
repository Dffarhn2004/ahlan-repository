package httpapi

import (
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/google/uuid"
)

// allowedOrigins returns the comma-separated CORS origins from CORS_ALLOWED_ORIGINS,
// defaulting to local dev plus the production frontend.
func allowedOrigins() string {
	if v := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS")); v != "" {
		return v
	}
	return "http://localhost:3000,https://ahlanapp.cloud,https://www.ahlanapp.cloud"
}

// newFiber builds the Fiber app with the shared error handler, CORS and request logging.
func newFiber() *fiber.App {
	app := fiber.New(fiber.Config{BodyLimit: 20 << 20, ReadBufferSize: 64 * 1024, ErrorHandler: func(c *fiber.Ctx, err error) error {
		status := http.StatusInternalServerError
		message := "INTERNAL_ERROR"
		if e, ok := err.(*fiber.Error); ok {
			status, message = e.Code, e.Message
		}
		requestID := c.GetRespHeader("X-Request-ID")
		slog.Error("http request failed", "method", c.Method(), "path", c.OriginalURL(), "status", status, "request_id", requestID, "error", err)
		return c.Status(status).JSON(fiber.Map{"code": message, "request_id": requestID})
	}})
	app.Use(cors.New(cors.Config{AllowOrigins: allowedOrigins(), AllowCredentials: true, AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS", AllowHeaders: "Origin,Content-Type,Accept,Authorization,X-Requested-With"}))
	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()
		id := uuid.NewString()
		c.Set("X-Request-ID", id)
		err := c.Next()
		slog.Info("http request", "method", c.Method(), "path", c.OriginalURL(), "status", c.Response().StatusCode(), "latency_ms", time.Since(start).Milliseconds(), "request_id", id)
		return err
	})
	app.Use(logger.New(logger.Config{Format: "${time} ${status} ${latency} ${method} ${path}\n"}))
	return app
}
