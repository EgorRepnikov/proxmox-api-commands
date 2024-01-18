FROM golang:1.20.13 AS build

WORKDIR /app

COPY go.mod go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o ./bin/proxmox-api-commands ./cmd/proxmox-api-commands

FROM alpine:3.19.0

WORKDIR /usr/bin

COPY --from=build /app/bin/proxmox-api-commands proxmox-api-commands

EXPOSE ${PORT}

ENTRYPOINT /usr/bin/proxmox-api-commands
