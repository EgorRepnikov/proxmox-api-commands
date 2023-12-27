all: build
build:
	GO111MODULE=on go build -o bin/proxmox-api-commands cmd/proxmox-api-commands/main.go
clean:
	rm -f bin/proxmox-api-commands
run:
	./bin/proxmox-api-commands
dev:
	make build && make run
