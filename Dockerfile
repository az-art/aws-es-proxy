#Builder
FROM golang:1.12-alpine as builder
RUN apk add --update bash curl git && \
    rm /var/cache/apk/* && \
    mkdir /go/src/app
ADD . /go/src/app
WORKDIR /go/src/app
RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v0.5.3/dep-linux-amd64 && chmod +x /usr/local/bin/dep
RUN dep ensure
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o aws-es-proxy ./cmd/server

#Runtime
FROM scratch
LABEL name="aws-es-proxy" \
      version="latest"

COPY --from=builder /go/src/app/aws-es-proxy /app/
WORKDIR /app
ENV PORT_NUM 9200
EXPOSE ${PORT_NUM}

ENTRYPOINT ["./aws-es-proxy"]
CMD ["-h"]