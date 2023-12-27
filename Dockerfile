FROM golang:latest AS build

ENV GO111MODULE=on

WORKDIR /go/src/app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/proxmox-api-commands ./cmd/proxmox-api-commands

# final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /usr/bin
COPY --from=build /go/src/app/bin/proxmox-api-commands proxmox-api-commands

EXPOSE ${PORT}
ENTRYPOINT /usr/bin/proxmox-api-commands