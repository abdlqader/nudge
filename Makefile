.PHONY: swagger dev build setup

## swagger: Regenerate OpenAPI spec from handler annotations
swagger:
	swag init -g main.go --output docs

## dev: Regenerate swagger spec then run the app
dev: swagger
	go run main.go

## build: Regenerate swagger spec then build binary
build: swagger
	go build -o bin/nudge .

## setup: Install swag CLI and configure the pre-commit git hook
setup:
	go install github.com/swaggo/swag/cmd/swag@latest
	git config core.hooksPath .githooks
	@echo "Setup complete. Run 'make swagger' to generate the initial spec."
