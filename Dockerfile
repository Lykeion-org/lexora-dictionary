FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/service ./cmd

FROM alpine:3.20
# The service loads its config from ../config.yaml (or ../dev_config.yaml) relative to the working directory
WORKDIR /app/cmd
COPY --from=build /out/service /app/service
ENTRYPOINT ["/app/service"]
