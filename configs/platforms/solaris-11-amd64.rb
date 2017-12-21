platform "solaris-11-amd64" do |plat|
  plat.vmpooler_template "solaris-11-x86_64"

  plat.install_build_dependencies_with "pkg install ", " || [[ $? -eq 4 ]]"
end
