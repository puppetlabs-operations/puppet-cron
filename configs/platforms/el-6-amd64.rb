platform "el-6-amd64" do |plat|
  plat.vmpooler_template "centos-6-x86_64"

  plat.provision_with "yum install --assumeyes createrepo rsync make rpmdevtools rpm-libs yum-utils rpm-sign"
end
