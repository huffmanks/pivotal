set dotenv-load

binary_name := "url-shortener"
image_repo := "huffmanks/url-shortener"
dist_path := "dist"
version := "1.0.0"
docker_builder := "url-shortener_builder"

default: dev

clean:
    rm -rf {{dist_path}}

dev *args:
    go run main.go {{args}}

build:
    goreleaser release --snapshot --clean

tag ver=version:
    git tag v{{ver}}
    git push origin v{{ver}}

release:
    GITHUB_TOKEN=$(gh auth token) goreleaser release --clean

docker-build push="false":
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