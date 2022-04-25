FROM golang:1.17-alpine AS build
RUN apk update
RUN apk upgrade
ADD . /app
WORKDIR /app
RUN CGO_ENABLED=0 go build -o webhook -a -installsuffix cgo .

FROM alpine
WORKDIR /
COPY --from=build /app/webhook /
ENTRYPOINT ["/webhook"]
