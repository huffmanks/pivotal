set dotenv-load

binary_name := "pivotal"
image_repo := "huffmanks/pivotal"
dist_path := "dist"
version := "1.0.0"
docker_builder := "pivotal_builder"

set default-list := true

# Clean build artifacts
clean:
    rm -rf {{dist_path}}

# Run both web/server concurrently
[parallel]
dev: server web

# Run Go server
server:
    @echo "🚀 Starting Server..."
    @go run ./cmd/server 2>&1 | awk '{print "\033[1;36m[SERVER]\033[0m " $0}'

# Run web
web:
    @echo "🌐 Starting Web Server..."
    @cd web && pnpm dev 2>&1 | awk '{print "\033[1;35m[WEB]\033[0m " $0}'

# Run all unit tests
test:
	go test -v -race -count=1 -coverpkg=./... -coverprofile=coverage.out ./...

# Generate coverage report
coverage: test
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Build server
build-server:
    goreleaser release --snapshot --clean

# Build web
build-web:
    @echo "🌐 Building Web..."
    @cd web && pnpm build

# Build server and web
build: build-web build-server

# Create tag
tag ver=version:
    git tag v{{ver}}
    git push origin v{{ver}}

# Release binary
release:
    GITHUB_TOKEN=$(gh auth token) goreleaser release --clean

# Build Docker image
build-docker push="false":
    @echo "Building docker image version {{version}}..."
    @docker buildx ls | grep -q {{docker_builder}} || docker buildx create --name {{docker_builder}} --driver docker-container
    docker buildx inspect --bootstrap
    docker buildx use {{docker_builder}}
    {{ if push == "true" { \
        "docker buildx build --platform linux/amd64,linux/arm64 " + \
        "-t " + image_repo + ":" + version + " " + \
        "-t " + image_repo + ":latest --push ." \
    } else { \
        "docker buildx build -t " + image_repo + ":local" + " --load ." \
    } }}
    @docker buildx rm {{docker_builder}} || true
    @echo "Build complete."