package embeddedswagger

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAddRegistersOpenAPIAndInitializerURL(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	cfg := Config{
		OpenAPI:    []byte("openapi-document"),
		OpenAPIURL: "/custom/openapi.json",
		SwaggerURL: "/docs/swagger",
	}

	if err := Add(cfg, mux); err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	t.Run("openapi endpoint", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/custom/openapi.json", nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}

		if got := rr.Body.String(); got != "openapi-document" {
			t.Fatalf("body = %q, want %q", got, "openapi-document")
		}
	})

	t.Run("swagger initializer uses configured openapi url", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/docs/swagger/swagger-initializer.js", nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}

		body := rr.Body.String()
		if !strings.Contains(body, "url: \"/custom/openapi.json\"") {
			t.Fatalf("initializer body = %q, want custom openapi url", body)
		}

		if strings.Contains(body, defaultSwaggerInitializerURLTemplate) {
			t.Fatalf("initializer body still references default petstore URL: %q", body)
		}
	})
}

func TestAddUsesDefaultPaths(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()

	cfg := DefaultConfig([]byte("default-doc"))
	if err := Add(cfg, mux); err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	cases := map[string]struct {
		path string
		want string
	}{
		"docs default": {
			path: "/docs",
			want: "default-doc",
		},
		"swagger initializer default": {
			path: "/swagger/swagger-initializer.js",
			want: "url: \"/docs\"",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, tc.path, nil)
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
			}

			if tc.want != "" && !strings.Contains(rr.Body.String(), tc.want) {
				t.Fatalf("body = %q, want substring %q", rr.Body.String(), tc.want)
			}
		})
	}
}
