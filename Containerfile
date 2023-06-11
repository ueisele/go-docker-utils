FROM golang:1.20.5-alpine as builder

RUN apk add --no-cache upx git \
    && rm -rf /var/cache/apk/*

WORKDIR /workspace
COPY . .
RUN go test -v ./...
RUN ./build.sh

FROM busybox:stable-musl
#FROM scratch # `scratch` is sufficient for godub
COPY --from=builder /workspace/godub /godub
ENTRYPOINT ["/godub"]