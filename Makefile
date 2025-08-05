SHELL := /usr/bin/env bash

INPUT ?=
OUTPUT ?= ./obfuscated_src
PROFILE ?= safe
LOG ?= info
DISABLE_ANTI_VM ?=

.PHONY: all build run obfuscate payload smoke clean profiles build-windows

all: build

build:
	@echo "==> building CLI"
	go build -ldflags="-s -w" -o obfuscator_cli .

run: build
	@echo "==> running CLI"
	@if [ -z "$(INPUT)" ]; then \
		echo "INPUT не задан. Пример: make run INPUT=./path/to/src"; \
		exit 1; \
	fi ;\
	DISABLE_FLAG=""; \
	if [ "$(DISABLE_ANTI_VM)" = "1" ]; then \
		DISABLE_FLAG="-disable-anti-vm"; \
		export OBF_DISABLE_ANTI_VM=1; \
	fi ;\
	./obfuscator_cli -input "$(INPUT)" -output "$(OUTPUT)" $DISABLE_FLAG

# Быстрые профили (информационные)
profiles: build
	@echo "==> profiles: $(PROFILE)"
	@case "$(PROFILE)" in \
		fast) \
			echo "RENAME=true ENCRYPT_STRINGS=true INSERT_DEAD_CODE=false OBF_FLOW=false OBF_EXPR=false OBF_DATA=false OBF_CONST=false ANTI_DEBUG=true ANTI_VM=true INDIRECT=false INTEGRITY=false META=false SELF_MOD=false" ;; \
		safe) \
			echo "RENAME=true ENCRYPT_STRINGS=true INSERT_DEAD_CODE=true OBF_FLOW=true OBF_EXPR=true OBF_DATA=true OBF_CONST=true ANTI_DEBUG=true ANTI_VM=true INDIRECT=true INTEGRITY=true META=true SELF_MOD=true" ;; \
		aggressive) \
			echo "RENAME=true ENCRYPT_STRINGS=true INSERT_DEAD_CODE=true OBF_FLOW=true OBF_EXPR=true OBF_DATA=true OBF_CONST=true ANTI_DEBUG=true ANTI_VM=true INDIRECT=true INTEGRITY=true META=true SELF_MOD=true" ;; \
		*) echo "[warn] неизвестный профиль $(PROFILE). Использую safe." ;; \
	esac

obfuscate: run

payload:
	@echo "==> running demo payload"
	PAYLOAD_PASSWORD=${PAYLOAD_PASSWORD:-} go run ./cmd/payload

smoke: build
	@echo "==> smoke: basic payload demo"
	PAYLOAD_PASSWORD= make -s payload || true

clean:
	@echo "==> clean"
	rm -f obfuscator_cli
	rm -rf "$(OUTPUT)"

# Windows cross-build
build-windows:
	@echo "==> building CLI for windows/amd64"
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o obfuscator_cli_windows.exe .