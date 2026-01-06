#!/bin/bash
# Copyright 2022-present Kuei-chun Chen. All rights reserved.
die() { echo "$*" 1>&2 ; exit 1; }

[[ "$(which go)" = "" ]] && die "go command not found"

GIT_DATE=$(git log -1 --date=format:"%Y%m%d" --format="%ad" 2>/dev/null || date +"%Y%m%d")
VERSION="$(cat version)-${GIT_DATE}"
REPO=$(basename "$(dirname "$(pwd)")")/$(basename "$(pwd)")
LDFLAGS="-X main.version=$VERSION -X main.repo=$REPO"
TAG="simagix/hatchet"

gover=$(go version | cut -d' ' -f3)
[[ "$gover" < "go1.21" ]] && die "go version 1.21 or above is recommended."

if [ ! -f go.sum ]; then
    go mod tidy
fi

print_usage() {
  echo "Usage: $0 [command]"
  echo ""
  echo "Commands:"
  echo "  (none)        Build binary for current platform to dist/"
  echo "  docker        Build Docker image for current platform (local)"
  echo "  push          Build and push multi-arch Docker image (amd64 + arm64)"
  echo "  binaries      Build binaries for all platforms (linux/mac, amd64/arm64)"
  echo ""
  echo "Internal (used by Dockerfile):"
  echo "  binary        Build binary for current platform (auto-detected)"
  echo ""
}

mkdir -p dist

case "$1" in
  docker)
    # Build for current platform only (local image for testing)
    BR=$(git branch --show-current)
    if [[ "${BR}" == "main" ]]; then
      BR=$(cat version)
    fi
    docker buildx build --load \
      -t ${TAG}:${BR} \
      -t ${TAG}:latest . || die "docker build failed"
    echo "Built ${TAG}:${BR} for $(uname -m)"
    ;;

  push)
    # Requires: docker buildx create --name multibuilder --driver docker-container --use
    BR=$(git branch --show-current)
    if [[ "${BR}" == "main" ]]; then
      BR=$(cat version)
    fi
    echo "Building multi-arch image for linux/amd64 and linux/arm64..."
    docker buildx build --builder multibuilder --platform linux/amd64,linux/arm64 \
      --provenance=false --sbom=false \
      -t ${TAG}:${BR} \
      -t ${TAG}:latest \
      --push . || die "docker build failed"
    echo "Pushed ${TAG}:${BR} (amd64 + arm64)"
    ;;

  binary)
    # Internal: called by Dockerfile. Cross-compiles using GOOS/GOARCH env vars
    LDFLAGS="${LDFLAGS} -X main.docker=docker"
    CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o hatchet main/hatchet.go
    echo "Built hatchet for ${GOOS:-$(go env GOOS)}/${GOARCH:-$(go env GOARCH)}"
    ;;

  binaries)
    echo "Building binaries for all platforms..."
    
    # Linux amd64
    CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/hatchet-linux-amd64 main/hatchet.go
    echo "  Built dist/hatchet-linux-amd64"
    
    # Linux arm64
    CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -ldflags "$LDFLAGS" -o dist/hatchet-linux-arm64 main/hatchet.go
    echo "  Built dist/hatchet-linux-arm64"
    
    # macOS amd64
    CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/hatchet-darwin-amd64 main/hatchet.go
    echo "  Built dist/hatchet-darwin-amd64"
    
    # macOS arm64 (Apple Silicon)
    CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o dist/hatchet-darwin-arm64 main/hatchet.go
    echo "  Built dist/hatchet-darwin-arm64"
    
    echo "Done! Binaries in dist/"
    ;;

  help|-h|--help)
    print_usage
    ;;

  "")
    rm -f ./dist/hatchet
    go build -ldflags "$LDFLAGS" -o ./dist/hatchet main/hatchet.go
    if [[ -f ./dist/hatchet ]]; then
      ./dist/hatchet -version
    fi
    ;;

  *)
    echo "Unknown command: $1"
    print_usage
    exit 1
    ;;
esac
