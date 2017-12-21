platform "el-7-amd64" do |plat|
  plat.vmpooler_template "centos-7-x86_64"

  plat.install_build_dependencies_with "yum install --assumeyes"
  plat.provision_with "yum install --assumeyes createrepo rsync make rpmdevtools rpm-libs yum-utils rpm-sign"
end
