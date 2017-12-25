#!/bin/sh

# Build for the platforms we care about. Sets up a clean GOPATH to build in.

set -e

build () {
  export GOOS=$1
  export GOARCH=$2
  echo "Building $GOOS-$GOARCH"
  mkdir -p "builds/$GOOS-$GOARCH/build"
  go build -o "builds/$GOOS-$GOARCH/build/puppet-cron"

  # Vanagon needs an archive rather than a plain file.
  (cd "builds/$GOOS-$GOARCH" && tar -czf build.tar.gz build)
}

export GOPATH="$(mktemp -d)"
echo "Building under $GOPATH"

mkdir "$GOPATH/src"
mkdir "$GOPATH/bin"
mkdir "$GOPATH/pkg"

curl https://glide.sh/get | sh

package="$(grep '^package: ' glide.yaml | cut -f 2- -d ' ')"

mkdir -p "$GOPATH/src/$(dirname "$package")"
ln -s "$(pwd)" "$GOPATH/src/${package}"

cd "$GOPATH/src/${package}"

rm -Rf builds

glide install

build darwin amd64
build linux amd64
build linux 386
build solaris amd64

cd
echo "Deleting $GOPATH"
rm -Rf "$GOPATH"
