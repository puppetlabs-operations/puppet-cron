#!/bin/sh

set -e

cd output/deb

# Ignore missing directories
mv wheezy sysops-wheezy || true
mv jessie sysops-jessie || true
mv stretch sysops-stretch || true

scp -qr sysops-* aptly@opsrepo-aptly1-prod.ops.puppetlabs.net:

echo
echo Published:
find sysops-* -type f
