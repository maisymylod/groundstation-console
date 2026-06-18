# Multi-stage build for the console backend. The frontend is built separately
# (frontend/) and served by any static host or the dev server; this image is the
# Go API the SPA reads from.
FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/console ./cmd/console

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/console /console
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/console"]
