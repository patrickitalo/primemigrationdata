.PHONY: dev build build-native

WAILS := $(shell go env GOPATH)/bin/wails

# Desenvolvimento: abre a janela com hot reload do frontend via Vite
dev:
	$(WAILS) dev

# Build de release para Windows (cross-compile do macOS requer MinGW)
build:
	$(WAILS) build -platform windows/amd64 -o dist/prime-migration-fyne.exe

# Build nativo (mesma plataforma)
build-native:
	$(WAILS) build -o dist/prime-migration-fyne
