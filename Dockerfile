FROM golang:1.16.6-alpine as builder

RUN apk add --no-cache gcc g++ upx git \
    && rm -rf /var/cache/apk/*

WORKDIR /workspace
COPY . .
RUN ./build-static.sh godub
RUN go test -v ./...

FROM alpine:3.14.0
RUN apk add --no-cache ca-certificates \
    && rm -rf /var/cache/apk/*
COPY --from=builder /workspace/godub /godub
CMD ["/godub"]