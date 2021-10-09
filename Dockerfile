FROM golang:1.17.2-alpine as builder

RUN apk add --no-cache gcc g++ upx git \
    && rm -rf /var/cache/apk/*

WORKDIR /workspace
COPY . .
RUN ./build-static.sh godub
RUN go test -v ./...

FROM busybox:stable-musl
COPY --from=builder /workspace/godub /godub
ENTRYPOINT ["/godub"]