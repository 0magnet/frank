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
