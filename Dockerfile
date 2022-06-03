FROM golang:1.17-alpine AS build
RUN apk update
RUN apk upgrade
ADD *go* /app/
WORKDIR /app
RUN CGO_ENABLED=0 go build -o webhook -a -installsuffix cgo .

FROM alpine
WORKDIR /usr/share/apm-mutating-webhook
COPY --from=build /app/webhook .
ADD LICENSE.txt NOTICE.txt /usr/share/apm-mutating-webhook/
ENTRYPOINT ["./webhook"]
