FROM golang:1.16.6-alpine as builder

RUN apk add --no-cache gcc g++ \
    && rm -rf /var/cache/apk/*

WORKDIR /workspace
COPY . .
RUN go build -o godub -tags 'netgo osusergo' -ldflags "-linkmode external -extldflags -static -X main.AppVersionMetadata=$(date -u +%s)"

FROM alpine:3.14.0
RUN apk add --no-cache ca-certificates \
    && rm -rf /var/cache/apk/*
COPY --from=builder /workspace/godub /godub
CMD ["/godub"]