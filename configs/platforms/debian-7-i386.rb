platform "debian-7-i386" do |plat|
  plat.codename "wheezy"
  plat.vmpooler_template "debian-7-i386"

  plat.provision_with <<-SCRIPT
    set -e
    export DEBIAN_FRONTEND=noninteractive
    apt-get update -qq
    apt-get install -qy --no-install-recommends make rsync curl devscripts fakeroot debhelper
  SCRIPT
end
