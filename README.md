# Wrapper for running the puppet agent from cron

Our workflow is to develop changes in feature branches, which become
environments. Once those branches are merged, they are deleted from the repo and
thus the environment is deleted.

This wrapper reverts to `production` if the configured environment is invalid.

### Usage

~~~
0,30 * * * * /usr/local/bin/puppet-cron
~~~

If there is an error, or if the environment is reset to `production`, the
wrapper will write to stdout, and thus generate an email.

## Requirements

### Runtime

  * `/opt/puppetlabs/bin/puppet`
  * `/var`

This does not directly use the gem or directly modify the puppet configuration
file. Everything goes through the puppet executable.

This can run on any [platforms] Go can target. Most relevant to us are Linux,
macOS, and Solaris.

### Build

  * A recent version of [Go] (1.9 will work; 1.7 might work)
  * [Glide]

Glide will install:

  * github.com/danielparks/lockfile
  * golang.org/x/sys

## Building

~~~
glide install
go build
~~~

You can skip `glide` if the dependencies haven't changed.


[platforms]: https://golang.org/doc/install/source#environment
[Go]: https://golang.org/
[Glide]: https://glide.sh
