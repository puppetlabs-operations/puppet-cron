platform "debian-8-i386" do |plat|
  plat.codename "jessie"
  plat.vmpooler_template "debian-8-i386"

  plat.provision_with <<-SCRIPT
    set -e
    export DEBIAN_FRONTEND=noninteractive
    apt-get update -qq
    apt-get install -qy --no-install-recommends make rsync curl devscripts fakeroot debhelper
  SCRIPT
end
