# syntax=docker/dockerfile:1.2
FROM golang:1.19-bullseye as builder
ENV DOCKER_BUILDKIT=1
#
WORKDIR /app
COPY go.mod ./
COPY . .
#
#COPY go.sum ./
#RUN go mod tidy
#RUN go mod graph | awk '{if ($1 !~ "@") print $2}' | xargs go get
RUN --mount=type=cache,target=/go/pkg/mod \
   --mount=type=cache,target=/root/.cache/go-build go mod tidy
#ARG VERSION
RUN --mount=type=cache,target=/go/pkg/mod \
   --mount=type=cache,target=/root/.cache/go-build \
   CGO_ENABLED=0 go build -installsuffix cgo -ldflags "-X main.version=1" -o ./evo .
#
#
#FROM phusion/baseimage:focal-1.2.0
#
#COPY --from=builder /app /app
#COPY --from=builder /app/cms /
#WORKDIR /app
#
EXPOSE 8080
#
CMD ["/build/evo", "-c","/build/config.yml"]
