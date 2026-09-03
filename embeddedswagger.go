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
	SwaggerVersion = "5.32.14"
	// DefaultSwaggerPattern is the default URL for Swagger-UI.
	DefaultSwaggerPattern = "/swagger"
	// DefaultOpenAPIPattern is the default URL for the OpenAPI specification.
	DefaultOpenAPIPattern                = "/docs"
	defaultSwaggerInitializerURLTemplate = "https://petstore.swagger.io/v2/swagger.json"
)

var (
	// ErrServerIsNil error indicating that the server parameter is nil.
	ErrServerIsNil = errors.New("server is nil")
	//go:embed static/swagger-ui/*
	swaggerUI embed.FS
)

type (
	// Config is the configuration for the package.
	Config struct {
		// OpenApi is the raw OpenAPI specification in bytes.
		OpenAPI []byte
		// OpenAPIURL is the URL of the OpenAPI specification. Default to '/docs'.
		OpenAPIURL string
		// SwaggerURL is the URL of the OpenAPI specification. Default to '/swagger'.
		SwaggerURL string
	}

	// ConfigError is an error that is thrown when the Config is not valid.
	ConfigError struct {
		Msg string
	}

	pattern string

	RouteRegistrar interface {
		Handle(pattern string, handler http.Handler)
	}

	RouteMethodRegistrar interface {
		Method(method, pattern string, handler http.Handler)
	}
)

// Add registers the Swagger endpoints on the provided mux.
func Add(cfg Config, s RouteRegistrar) error {
	if s == nil {
		return ErrServerIsNil
	}

	if err := cfg.Validate(); err != nil {
		return err
	}

	openAPIPath := cfg.openAPIPath()
	swaggerURL := cfg.swaggerPath()
	initialContent, err := swaggerInitializerSource(normalizeURLPath(openAPIPath.pattern()))
	if err != nil {
		return err
	}

	registerGetRoute(s, openAPIPath.pattern(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType, hasContentType := openAPIPath.contentType()
		if hasContentType {
			w.Header().Set("Content-Type", contentType)
		}

		_, _ = w.Write(cfg.OpenAPI)
	}))

	registerGetRoute(s, fmt.Sprintf("%s/swagger-initializer.js", swaggerURL), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		_, _ = w.Write(initialContent)
	}))

	sfs, err := fs.Sub(swaggerUI, "static/swagger-ui")
	if err != nil {
		return fmt.Errorf("failed to load swagger-ui: %w", err)
	}

	fsHandler := http.StripPrefix(swaggerURL.pattern(), http.FileServer(http.FS(sfs)))
	registerGetRoute(s, staticSwaggerPattern(s, swaggerURL.pattern()), fsHandler)

	return nil
}

func registerGetRoute(s RouteRegistrar, routePattern string, handler http.Handler) {
	if methodRegistrar, ok := s.(RouteMethodRegistrar); ok {
		methodRegistrar.Method(http.MethodGet, routePattern, handler)

		return
	}

	s.Handle(routePattern, handler)
}

func staticSwaggerPattern(s RouteRegistrar, swaggerPath string) string {
	if _, ok := s.(RouteMethodRegistrar); ok {
		return fmt.Sprintf("%s/*", swaggerPath)
	}

	return fmt.Sprintf("%s/", swaggerPath)
}

// Error implements the error interface.
func (e *ConfigError) Error() string {
	return e.Msg
}

func DefaultConfig(openapi []byte) Config {
	return Config{
		OpenAPI:    openapi,
		OpenAPIURL: DefaultOpenAPIPattern,
		SwaggerURL: DefaultSwaggerPattern,
	}
}

// Validate checks that the Config has correct values.
func (c *Config) Validate() error {
	if len(c.OpenAPI) == 0 {
		return &ConfigError{Msg: "OpenAPI specification is empty"}
	}
	if c.OpenAPIURL == "" {
		return &ConfigError{Msg: "OpenAPI URL is empty"}
	}
	if c.SwaggerURL == "" {
		return &ConfigError{Msg: "Swagger URL is empty"}
	}
	return nil
}

func (c *Config) openAPIPath() pattern {
	if c.OpenAPIURL == "" {
		return DefaultOpenAPIPattern
	}

	return pattern(normalizeURLPath(c.OpenAPIURL))
}

func (c *Config) swaggerPath() pattern {
	if c.SwaggerURL == "" {
		return DefaultSwaggerPattern
	}

	return pattern(normalizeURLPath(c.SwaggerURL))
}

func (p pattern) pattern() string {
	return string(p)
}

func (p pattern) contentType() (string, bool) {
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
