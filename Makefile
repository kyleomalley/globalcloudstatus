.PHONY: worker worker-watch web build dev

OUTPUT ?= web/public/data/status.json

# Run the worker once to populate status.json
worker:
	go run ./cmd/worker -output $(OUTPUT)

# Run the worker in watch mode (refreshes every 5 min by default)
worker-watch:
	go run ./cmd/worker -output $(OUTPUT) -watch

# Start the Vite dev server
web:
	cd web && npm install && npm run dev

# Build the production web bundle
web-build:
	cd web && npm run build

# Build the worker binary
build:
	go build -o bin/worker ./cmd/worker

# Run worker once, then start dev server
dev:
	$(MAKE) worker
	$(MAKE) web
