package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/manuelarte/embeddedswagger"
)

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
		"mux server": {
			httpServer:  http.NewServeMux(),
			openapipath: "/docs",
		},
		"chi server": {
			httpServer:  chi.NewRouter(),
			openapipath: "/docs",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if err := embeddedswagger.Add(embeddedswagger.Config{
				OpenAPI: OpenAPI,
			}, tc.httpServer); err != nil {
				t.Fatalf("Add returned error: %v", err)
			}

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, tc.openapipath, nil)
			rr := httptest.NewRecorder()
			tc.httpServer.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
			}
			if got := rr.Body.String(); got != string(OpenAPI) {
				t.Fatalf("body = %q, want %q", got, string(OpenAPI))
			}
		})
	}
}
