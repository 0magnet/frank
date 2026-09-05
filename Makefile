# Frank — native browser on WebKitGTK-6.0 (GTK4).
# Requires system libraries: gtk4, webkitgtk-6.0

run:
	@go run .

build:
	@go build -o frank .

test:
	@go test ./...

vet:
	@go vet ./...

clean:
	@rm -f frank

test-wasm: ## Compile-gate the js/wasm-tagged half
	@# Self-contained: this Makefile does not carry the shared template's
	@# PKGS/JSPKGS variables. A host build cannot see //go:build js && wasm
	@# files at all, so without this a wasm-only break stays invisible.
	@#
	@# A BUILD, not a test run: a package can support js/wasm while its tests
	@# do not — bbolt's unix_test.go needs unix.Rlimit, which does not exist
	@# there. Compiling is what catches the error this target exists for.
	@if ! grep -rlq '^//go:build js' --include='*.go' --exclude-dir=vendor . 2>/dev/null; then \
		echo 'no js/wasm-tagged files; nothing to gate'; \
	else \
		echo '--- building in the js/wasm build context'; \
		CGO_ENABLED=0 GOOS=js GOARCH=wasm go build ./...; \
	fi

lint: ## Run golangci-lint, in the host context and again for js/wasm
	command -v golangci-lint || go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	golangci-lint run
	@# A host run cannot see js/wasm-tagged files, so anything wrong inside
	@# them is never checked at all.
	@if grep -rlq '^//go:build js' --include='*.go' --exclude-dir=vendor . 2>/dev/null; then \
		echo '--- again in the js/wasm build context'; \
		CGO_ENABLED=0 GOOS=js GOARCH=wasm golangci-lint run; \
	fi
