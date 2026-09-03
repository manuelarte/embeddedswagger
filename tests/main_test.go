package tests

import (
	"testing"

	"github.com/manuelarte/embeddedswagger"
	root "github.com/manuelarte/embeddedswagger/tests"
)

func TestOpenAPIPath(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		createServerFn func()
		openapipath    string
	}{
		"mux server": {},
		"chi server": {},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s := tc.createServerFn()
			embeddedswagger.Add(embeddedswagger.Config{
				OpenAPI: root.OpenAPI,
			}, s)

			// Check /docs endpoint that it's 200.

		})
	}
}
