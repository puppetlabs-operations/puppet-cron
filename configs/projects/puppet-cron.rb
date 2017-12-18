project "puppet-cron" do |proj|
  proj.setting(:prefix, "/opt/puppet-cron")
  proj.setting(:bindir, File.join(proj.prefix, "bin"))
  proj.setting(:piddir, "/var/run")

  proj.description "Wrapper for puppet agent to be run from cron"
  proj.version "1.0.0"
  proj.homepage "https://github.com/puppetlabs-operations/puppet-cron"
  proj.vendor "Daniel Parks <daniel.parks@puppet.com>"
  proj.license "MIT"

  if proj.get_platform().is_solaris?
    proj.identifier "puppetlabs.com"
  elsif proj.get_platform().is_macos?
    proj.identifier "com.puppetlabs"
  end

  proj.directory proj.prefix
end
