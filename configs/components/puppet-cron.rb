component "puppet-cron" do |pkg, settings, platform|
  pkg.version "1.0.0"
  pkg.url = "git@github.com:puppetlabs-operations/puppet-cron.git"
  pkg.license "MIT"

  # Figure out URL to download Go from
  arch = platform.architecture
  if arch == "i386"
    arch = "386"
  end

  if platform.is_linux?
    os = "linux"
  elsif platform.is_macos?
    os = "darwin"
  else
    os = platform.os_name
  end

  go_url = "https://redirector.gvt1.com/edgedl/go/go1.9.2.#{os}-#{arch}.tar.gz"

  # Actual build stuff
  pkg.configure do
    [
      "curl -sSL #{go_url} | tar -C /usr/local -xzf -",
      "for p in /usr/local/go/bin/* ; do ln -s $p /usr/bin ; done",
      "curl -sSL https://glide.sh/get | sh",
      "glide install",
    ]
  end

  pkg.build do
    ["go build"]
  end

  pkg.install do
    [
      "mkdir -p #{settings[:bindir]}",
      "cp puppet-cron #{settings[:bindir]}/"
    ]
  end
end
