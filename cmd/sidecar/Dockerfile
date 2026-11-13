FROM golang:1 AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /out/sidecar ./cmd/sidecar

FROM scratch
COPY --from=build /out/sidecar /sidecar
ENTRYPOINT ["/sidecar"]