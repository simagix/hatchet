#! /bin/bash
# Copyright 2022-present Kuei-chun Chen. All rights reserved.
# build.sh

die() { echo "$*" 1>&2 ; exit 1; }
VERSION="v$(cat version)-$(git log -1 --date=format:"%Y%m%d" --format="%ad")"
REPO=$(basename "$(dirname "$(pwd)")")/$(basename "$(pwd)")
LDFLAGS="-X main.version=$VERSION -X main.repo=$REPO"
TAG="simagix/hatchet"
[[ "$(which go)" = "" ]] && die "go command not found"

gover=$(go version | cut -d' ' -f3)
if [ "$gover" \< "go1.18" ]; then
    [[ "$GOPATH" = "" ]] && die "GOPATH not set"
    [[ "${GOPATH}/src/github.com/$REPO" != "$(pwd)" ]] && die "building hatchet should be under ${GOPATH}/src/github.com/$REPO"
fi

if [ ! -f go.sum ]; then
    go mod tidy
fi

mkdir -p dist
if [ "$1" == "docker" ]; then
  BR=$(git branch --show-current)
  if [[ "${BR}" == "main" ]]; then
    BR="latest"
  fi
  docker build --no-cache -f Dockerfile -t ${TAG}:${BR} .
  # docker rmi -f $(docker images -f "dangling=true" -q) > /dev/null 2>&1
elif [ "$1" == "dist" ]; then
  [[ "$(which uname)" = "" ]] && die "uname command not found"
  ofile="./dist/hatchet-$(uname|tr '[:upper:]' '[:lower:]')-$(uname -m)"
  go build -ldflags "$LDFLAGS" -o ${ofile} main/hatchet.go
else
  rm -f ./dist/hatchet
  go build -ldflags "$LDFLAGS" -o ./dist/hatchet main/hatchet.go
  if [[ -f ./dist/hatchet ]]; then
    ./dist/hatchet -version
  fi
fi
