project "puppet-cron" do |proj|
  proj.setting(:prefix, "/opt/puppet-cron")
  proj.setting(:bindir, File.join(proj.prefix, "bin"))
  proj.setting(:piddir, "/var/run")

  working_dir = File.dirname(File.dirname(File.dirname(File.realpath(__FILE__))))
  proj.setting(:working_dir, working_dir)

  git_version = ::Git.open(working_dir).describe("HEAD", tag: true, always: true)
  proj.version git_version
  proj.setting(:git_version, git_version)

  proj.description "Wrapper for puppet agent to be run from cron"
  proj.homepage "https://github.com/puppetlabs-operations/puppet-cron"
  proj.vendor "Daniel Parks <daniel.parks@puppet.com>"
  proj.license "MIT"

  if proj.get_platform().is_solaris?
    proj.identifier "puppetlabs.com"
  elsif proj.get_platform().is_macos?
    proj.identifier "com.puppetlabs"
  end

  proj.directory proj.prefix
  proj.component "puppet-cron"
end
