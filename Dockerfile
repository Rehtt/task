FROM golang:1.20-alpine3.17 AS build

RUN mkdir /tmp/build
WORKDIR /tmp/build
COPY a.tar .

RUN tar -xf a.tar && \
    export GOPROXY="https://goproxy.cn" && \
    go mod tidy && \
    go build -o exec


FROM alpine:3.17

COPY --from=build /tmp/build/exec .
EXPOSE 8080
CMD ["./exec"]
