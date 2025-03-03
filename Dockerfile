ARG GO_VERSION=1.24
FROM golang:${GO_VERSION}-alpine AS build
ADD *go* /app/
WORKDIR /app
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s" -o webhook .

FROM cgr.dev/chainguard/static:latest@sha256:853bfd4495abb4b65ede8fc9332513ca2626235589c2cef59b4fce5082d0836d
WORKDIR /usr/share/apm-k8s-attacher
COPY --from=build /app/webhook .
ADD LICENSE.txt NOTICE.txt /usr/share/apm-k8s-attacher/
ENTRYPOINT ["./webhook"]
