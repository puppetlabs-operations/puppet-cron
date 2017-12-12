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