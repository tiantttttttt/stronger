FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/stronger-api ./cmd/api

FROM alpine:3.22
RUN addgroup -S stronger && adduser -S stronger -G stronger
USER stronger
COPY --from=build /out/stronger-api /usr/local/bin/stronger-api
EXPOSE 8080
ENTRYPOINT ["stronger-api"]
