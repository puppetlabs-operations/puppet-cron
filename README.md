# Wrapper for running the puppet agent from cron

Our workflow is to develop changes in feature branches, which become
environments. Once those branches are merged, they are deleted from the repo and
thus the environment is deleted.

This wrapper reverts to `production` if the configured environment is invalid.

This does not directly use the gem or directly modify the puppet configuration
file. Everything goes through the puppet executable.

This can run on any [platforms] [Go] can target. Most relevant to us are Linux,
macOS, Solaris, and Windows.

## Usage

### Linux, macOS, Solaris x86

Assuming you have installed this to `/usr/local/bin/puppet-cron`, create a cron job similar
to the one below. The reason for the path information is so that facts will be resolved
correctly.

```bash
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/puppetlabs/bin
0,30 * * * * /usr/local/bin/puppet-cron
```

If there is an error, or if the environment is reset to `production`, the
wrapper will write to stdout, and thus generate an email.

### Windows

Configure a scheduled task to run `puppet-cron` every thirty minutes.

## Building

- See `go.mod` for a known working version of Go to use with this application.
- See `build.sh` for how the application is built. This script is designed to work both locally and in CI.

[platforms]: https://golang.org/doc/install/source#environment
[Go]: https://golang.org/
