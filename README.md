# Embedded Swagger

Simple Swagger UI setup for Go's standard library HTTP server.

## Install

```bash
go get github.com/manuelarte/embeddedswagger
```

## Use

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
	}, http.DefaultServeMux); err != nil {
		panic(err)
	}

	http.ListenAndServe(":8080", nil)
}
```

Check the [example](example) for a complete example.

## Defaults

- OpenAPI: `/docs`
- Swagger UI: `/swagger`
