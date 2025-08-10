SWAG_DIR=./
SWAG_OUTPUT=./docs

.PHONY: swag
swag:
	swag init --generalInfo ./cmd/main.go --dir ./internal/api --output ./docs --parseDependency=false
