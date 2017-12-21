component "puppet-cron" do |pkg, settings, platform|
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

  pkg.version settings[:git_version]
  pkg.url settings[:working_dir]
  pkg.license "MIT"

  # There is a bug (I think) in Vanagon that prevents using a local file as the
  # primary source (url) of a component. I set the URL to the local repo
  # (because it has to be set to something, and it might as well be something
  # local), then I add the build tarball as second source.
  pkg.add_source(File.join(settings[:working_dir], "builds/#{os}-#{arch}/build.tar.gz"))

  pkg.directory settings[:bindir]

  # The build tarball unpacks in the parent directory.
  pkg.install_file "../build/puppet-cron", File.join(settings[:bindir], "puppet-cron"),
    mode: "0755"
end
