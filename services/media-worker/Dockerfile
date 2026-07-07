FROM golang:1.24-alpine AS build
WORKDIR /src
COPY services/media-worker/go.mod services/media-worker/go.sum* ./services/media-worker/
RUN cd services/media-worker && go mod download
COPY services/media-worker ./services/media-worker
RUN cd services/media-worker && CGO_ENABLED=0 go build -trimpath -o /out/media-worker ./cmd/worker

FROM alpine:3.21
RUN apk add --no-cache ca-certificates ffmpeg
COPY --from=build /out/media-worker /media-worker
RUN addgroup -S cineforge && adduser -S -G cineforge -h /nonexistent cineforge
USER cineforge
ENTRYPOINT ["/media-worker"]
