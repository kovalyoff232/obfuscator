SHELL := /usr/bin/env bash


INPUT ?=
OUTPUT ?= ./obfuscated_src
PROFILE ?= safe
LOG ?= info

.PHONY: all build run obfuscate payload smoke clean

all: build

build:
	@echo "==> building CLI"
	go build -ldflags="-s -w" -o obfuscator_cli .

run: build
	@echo "==> running CLI"
	@if [ -z "$(INPUT)" ]; then \
		echo "INPUT не задан. Пример: make run INPUT=./path/to/src"; \
		exit 1; \
	fi
	./obfuscator_cli -input "$(INPUT)" -output "$(OUTPUT)" -profile "$(PROFILE)" -log "$(LOG)"

obfuscate: run

payload:
	@echo "==> running demo payload"
	PAYLOAD_PASSWORD=$${PAYLOAD_PASSWORD:-} go run ./cmd/payload

smoke: build
	@echo "==> smoke: basic payload demo"
	PAYLOAD_PASSWORD= make -s payload || true

clean:
	@echo "==> clean"
	rm -f obfuscator_cli
	rm -rf "$(OUTPUT)"