FROM golang:1.24-alpine AS build
WORKDIR /src
COPY services/control-api/go.mod services/control-api/go.sum* ./services/control-api/
RUN cd services/control-api && go mod download
COPY services/control-api ./services/control-api
RUN cd services/control-api && CGO_ENABLED=0 go build -trimpath -o /out/control-api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/control-api /control-api
USER nonroot:nonroot
ENTRYPOINT ["/control-api"]
