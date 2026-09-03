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

Check the [examples](examples) folder for a complete example on:

- Standard library `http.DefaultServeMux`.
- Chi router.
