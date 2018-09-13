#!/bin/sh

set -e

# PATH doesn't seem to include /usr/local/bin by default
export PATH="/usr/local/bin:$PATH"

bundle install --deployment

for a in configs/projects/*.rb ; do
  project=$(basename ${a%.rb})

  for p in configs/platforms/*.rb ; do
    platform=$(basename ${p%.rb})
    echo
    echo ======================================================================
    echo Packaging $project for $platform
    echo ======================================================================
    bundle exec build "${project}" "${platform}"
  done
done
