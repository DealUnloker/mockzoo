// Package docs embeds the hand-written OpenAPI specification (the single source of
// truth for the /v1 surface) and the static API reference page that renders it.
// Regenerating these from code is deliberately not supported: any handler change
// must update openapi.json in the same commit.
package docs

import _ "embed"

//go:embed openapi.json
var OpenAPIJSON []byte

//go:embed index.html
var ReferenceHTML []byte
