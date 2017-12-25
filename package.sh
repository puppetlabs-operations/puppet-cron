#!/bin/sh

set -e

bundle install --deployment

for p in configs/platforms/*.rb ; do
  platform=$(basename ${p%.rb})
  echo Packaging for platform
  bundle exec build puppet-cron "${platform}"
done
