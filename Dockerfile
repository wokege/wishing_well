FROM golang:1.22.0-alpine3.19 as build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /app/wishing ./...

FROM alpine:3.19.1 as run
WORKDIR /app
COPY --from=build /app/wishing /app/wishing
ENTRYPOINT /app/wishing