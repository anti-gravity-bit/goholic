.PHONY: test run build tidy static

static:
	go run ./cmd/goholic-static

tidy:
	go mod tidy

test:
	go test ./...

build:
	go build -o goholic-web-server ./cmd/goholic-web-server

run:
	GOHOLIC_WRITER_USERNAME=abir \
	GOHOLIC_WRITER_PASSWORD=change-me-now \
	GOHOLIC_SESSION_SECRET=change-this-secret! \
	go run ./cmd/goholic-web-server
