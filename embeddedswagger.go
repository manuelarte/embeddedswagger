// Package embeddedswagger contains the means to add Swagger-UI to your HTTP server.
package embeddedswagger

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
)

const (
	// SwaggerVersion is the version of Swagger-UI used.
	SwaggerVersion                       = "5.32.14"
	defaultSwaggerInitializerURLTemplate = "https://petstore.swagger.io/v2/swagger.json"
)

var (
	//go:embed static/swagger-ui/*
	swaggerUI embed.FS
	// ErrMuxIsNil error indicating that the mux parameter is nil.
	ErrMuxIsNil = errors.New("mux is nil")
)

type (
	// Config is the configuration for the package.
	Config struct {
		// OpenApi is the raw OpenAPI specification in bytes.
		OpenAPI []byte
		// OpenAPIURL is the URL of the OpenAPI specification. Default to /docs
		OpenAPIURL string
		// SwaggerURL is the URL of the OpenAPI specification. Default to /swagger
		SwaggerURL string
	}

	// InvalidOpenAPIError is an error that is thrown when the OpenAPI specification is not valid.
	InvalidOpenAPIError struct {
		Msg string
	}

	pattern     string
	swaggerPath pattern
	openapiPath pattern
)

// Add registers the Swagger endpoints on the provided mux.
func Add(cfg Config, mux *http.ServeMux) error {
	if mux == nil {
		return ErrMuxIsNil
	}

	if err := cfg.Validate(); err != nil {
		return err
	}

	openAPIPath := cfg.openAPIPath()
	mux.HandleFunc(openAPIPath.pattern(), func(w http.ResponseWriter, r *http.Request) {
		contentType, hasContentType := openAPIPath.contentType()
		if hasContentType {
			w.Header().Set("Content-Type", contentType)
		}

		_, _ = w.Write(cfg.OpenAPI)
	})

	swaggerURL := cfg.swaggerPath()

	initialContent, err := swaggerInitializerSource(normalizeURLPath(openAPIPath.pattern()))
	if err != nil {
		return err
	}

	mux.HandleFunc(fmt.Sprintf("%s/swagger-initializer.js", swaggerURL), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		_, _ = w.Write(initialContent)
	})

	sfs, err := fs.Sub(swaggerUI, "static/swagger-ui")
	if err != nil {
		return fmt.Errorf("failed to load swagger-ui: %w", err)
	}

	fsHandler := http.StripPrefix(swaggerURL.pattern(), http.FileServer(http.FS(sfs)))
	mux.Handle(fmt.Sprintf("%s/", swaggerURL.pattern()), fsHandler)

	return nil
}

// Error implements the error interface.
func (e *InvalidOpenAPIError) Error() string {
	return e.Msg
}

// Validate checks that the Config has correct values.
func (c *Config) Validate() error {
	if len(c.OpenAPI) == 0 {
		return &InvalidOpenAPIError{Msg: "OpenAPI specification is empty"}
	}
	// validate everything
	return nil
}

func (c *Config) openAPIPath() openapiPath {
	if c.OpenAPIURL == "" {
		return "/docs"
	}

	return openapiPath(normalizeURLPath(c.OpenAPIURL))
}

func (c *Config) swaggerPath() swaggerPath {
	if c.SwaggerURL == "" {
		return "/swagger"
	}

	return swaggerPath(normalizeURLPath(c.SwaggerURL))
}

func (p swaggerPath) pattern() string {
	return string(p)
}

func (p openapiPath) pattern() string {
	return string(p)
}

func (p openapiPath) contentType() (string, bool) {
	switch {
	case strings.HasSuffix(string(p), ".json"):
		return "application/json", true
	case strings.HasSuffix(string(p), ".yaml"), strings.HasSuffix(string(p), ".yml"):
		return "application/yaml", true
	default:
		return "", false
	}
}

func normalizeURLPath(raw string) string {
	if raw == "" {
		return "/"
	}

	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "/") {
		raw = fmt.Sprintf("/%s", raw)
	}

	if raw == "/" {
		return raw
	}

	return strings.TrimRight(raw, "/")
}

func swaggerInitializerSource(openAPIURL string) ([]byte, error) {
	sfs, err := fs.Sub(swaggerUI, "static/swagger-ui")
	if err != nil {
		return nil, fmt.Errorf("failed to load swagger-ui: %w", err)
	}

	content, err := fs.ReadFile(sfs, "swagger-initializer.js")
	if err != nil {
		return nil, fmt.Errorf("failed to read swagger-initializer.js: %w", err)
	}

	updated := strings.Replace(string(content), defaultSwaggerInitializerURLTemplate, openAPIURL, 1)

	return []byte(updated), nil
}
