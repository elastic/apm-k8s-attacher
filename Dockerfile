ARG GO_VERSION=1.22
FROM golang:${GO_VERSION}-alpine AS build
RUN apk update
RUN apk upgrade
ADD *go* /app/
WORKDIR /app
RUN CGO_ENABLED=0 go build -o webhook -a -installsuffix cgo .

FROM alpine
WORKDIR /usr/share/apm-k8s-attacher
COPY --from=build /app/webhook .
ADD LICENSE.txt NOTICE.txt /usr/share/apm-k8s-attacher/
ENTRYPOINT ["./webhook"]
