platform "el-6-i386" do |plat|
  plat.vmpooler_template "centos-6-i386"

  plat.provision_with "yum install --assumeyes createrepo rsync make rpmdevtools rpm-libs yum-utils rpm-sign"
end
