FROM golang:1.22-bookworm AS build_the_goholic_binary
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /goholic-web-server ./cmd/goholic-web-server

FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=build_the_goholic_binary /goholic-web-server /app/goholic-web-server
ENV GOHOLIC_LISTEN_ADDRESS=:8080
ENV GOHOLIC_DATABASE_FILE_PATH=/app/data/goholic.sqlite
EXPOSE 8080
USER nonroot:nonroot
VOLUME ["/app/data"]
ENTRYPOINT ["/app/goholic-web-server"]
