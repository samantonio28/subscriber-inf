package main

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config configs/backend.yaml doc.yaml

// Generate types and server for internal API
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -generate types -o internal/api/openapi_types.gen.go -package api doc.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -generate gorilla -o internal/api/openapi_api.gen.go -package api doc.yaml

//go:generate mkdir -p pkg/clients/api
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -generate types -o pkg/clients/api/openapi_types.gen.go -package api doc.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -generate client -o pkg/clients/api/openapi_client_gen.go -package api doc.yaml

// Generate types and server for v2 API (openapi.yaml, OpenAPI 3.1)
//go:generate mkdir -p internal/api/v2
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -generate types -o internal/api/v2/openapi_types.gen.go -package v2 openapi.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -generate gorilla -o internal/api/v2/openapi_api.gen.go -package v2 openapi.yaml

//go:generate mkdir -p pkg/clients/apiv2
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -generate types -o pkg/clients/apiv2/openapi_types.gen.go -package apiv2 openapi.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -generate client -o pkg/clients/apiv2/openapi_client_gen.go -package apiv2 openapi.yaml
