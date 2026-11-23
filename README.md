# Embedded Swagger

Simple Swagger UI setup (no dependencies) for Go HTTP server.

## Install

```bash
go get github.com/manuelarte/embeddedswagger
```

## Use

Example use for the standard library `http.DefaultServeMux`:

```go
package main

import (
	"net/http"

	"github.com/manuelarte/embeddedswagger"
)

func main() {
	openAPI := []byte(`{"openapi":"3.0.0"}`)

	if err := embeddedswagger.Add(embeddedswagger.Config{
		OpenAPI: openAPI,
		OpenAPIURL: "/docs",
		SwaggerURL: "/swagger",
	}, http.DefaultServeMux); err != nil {
		panic(err)
	}

	http.ListenAndServe(":8080", nil)
}
```

Then swagger will be available at `http://localhost:8080/swagger/`.

## Examples

Check the [examples](examples) folder for a complete example on:

- Standard library `http.DefaultServeMux`.
- Chi router.

## Updating Swagger UI

The weekly update workflow checks the latest Swagger UI release and opens or
updates a pull request containing its `dist` directory and the matching
`SwaggerVersion`. Run `make update-swagger-ui` to perform the same update
locally. It uses the complete `dist` directory as packaged in the upstream
source archive.

The repository must allow GitHub Actions to create pull requests. The workflow
runs the Go tests before opening a pull request; pull requests created with
`GITHUB_TOKEN` do not automatically trigger the other pull request workflows.
To trigger them, set the `SWAGGER_UI_UPDATE_TOKEN` repository secret to a
personal access token with contents and pull request write permissions.

The updater requires verified HTTPS. If Python cannot find a CA bundle on macOS,
run it with `SSL_CERT_FILE=/etc/ssl/cert.pem make update-swagger-ui`.
