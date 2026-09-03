package tests

import (
	_ "embed"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/manuelarte/embeddedswagger"
)

//go:embed openapi.json
var openAPI []byte

type httpServerFramework interface {
	embeddedswagger.RouteRegistrar
	http.Handler
}

func TestOpenAPIPath(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		httpServer  httpServerFramework
		openapipath string
	}{
		"mux server, empty (default) path": {
			httpServer: http.NewServeMux(),
		},
		"mux server, default path": {
			httpServer:  http.NewServeMux(),
			openapipath: "/docs",
		},
		"chi server, empty (default) path": {
			httpServer: chi.NewRouter(),
		},
		"chi server, default path": {
			httpServer:  chi.NewRouter(),
			openapipath: "/docs",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cfg := embeddedswagger.DefaultConfig(openAPI)
			if tc.openapipath != "" {
				cfg.OpenAPIURL = tc.openapipath
			}
			if err := embeddedswagger.Add(cfg, tc.httpServer); err != nil {
				t.Fatalf("Add returned error: %v", err)
			}

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, cfg.OpenAPIURL, nil)
			rr := httptest.NewRecorder()
			tc.httpServer.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
			}
			if got := rr.Body.String(); got != string(openAPI) {
				t.Errorf("body = %q, want %q", got, string(openAPI))
			}
		})
	}
}

func TestSwaggerPath(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		httpServer  httpServerFramework
		swaggerPath string
		validPaths  []string
	}{
		"mux server, empty (default) path": {
			httpServer: http.NewServeMux(),
			validPaths: []string{"/swagger", "/swagger/", "/swagger/index.html"},
		},
		"chi server, empty (default) path": {
			httpServer: chi.NewRouter(),
			validPaths: []string{"/swagger", "/swagger/", "/swagger/index.html"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cfg := embeddedswagger.DefaultConfig(openAPI)
			if tc.swaggerPath != "" {
				cfg.SwaggerURL = tc.swaggerPath
			}
			if err := embeddedswagger.Add(cfg, tc.httpServer); err != nil {
				t.Fatalf("Add returned error: %v", err)
			}

			for _, path := range tc.validPaths {
				req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
				rr := httptest.NewRecorder()
				tc.httpServer.ServeHTTP(rr, req)

				if rr.Code != http.StatusOK {
					t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
				}
				if got := rr.Body.String(); strings.Contains(got, "url: \"https://petstore.swagger.io/v2/swagger.json\",") {
					t.Errorf("body = %q, want %q", got, string(openAPI))
				}
			}

		})
	}
}
