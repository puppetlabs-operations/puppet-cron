#!/usr/bin/env bash

set -e

release_path="$(pwd)/.release"
repo_name="$(grep '^module' go.mod |cut -d ' ' -f2 |rev |cut -d '/' -f1 |rev)"
targets=${@-"aix/ppc64 darwin/amd64 darwin/arm64 linux/amd64 linux/386 solaris/amd64 windows/amd64 windows/386"}

echo "----> Setting up Go repository"
rm -rf ${release_path}
mkdir -p ${release_path}

for target in $targets; do
  os="$(echo $target | cut -d '/' -f1)"
  arch="$(echo $target | cut -d '/' -f2)"

  if [ "$os" = "windows" ]; then
    output="${release_path}/${repo_name}_${os}_${arch}.exe"
  else
    output="${release_path}/${repo_name}_${os}_${arch}"
  fi

  echo "----> Building project for: $target"
  GOOS=$os GOARCH=$arch CGO_ENABLED=0 go build -o $output

  zip -j $output.zip $output > /dev/null 2>&1
  tar -czvf $output.tgz $output > /dev/null 2>&1

  sha256sum $output |cut -d ' ' -f1 > $output-sha256sum.txt
  sha256sum $output.zip |cut -d ' ' -f1 > $output-zip-sha256sum.txt
  sha256sum $output.tgz |cut -d ' ' -f1 > $output-tgz-sha256sum.txt
done

echo "----> Build is complete. List of files at $release_path:"
cd $release_path
ls -al
