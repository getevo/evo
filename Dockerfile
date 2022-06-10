# Multi stage build
FROM golang:1.17 as builder
ENV GO111MODULE=off
#RUN go get -d github.com/getevo/evo

WORKDIR /build
COPY . .
RUN go get -d ./...
RUN go build -o main .

# Only runtime
FROM golang:1.17
COPY --from=builder /build/main /build/main
EXPOSE 8080
CMD ["/build/main","-c","/build/config.yml"]
