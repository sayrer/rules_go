workspace(name = "io_bazel_rules_go")

load("//go:def.bzl", "go_repositories", "go_internal_tools_deps", "new_go_vendor")

go_repositories()

new_go_vendor(
    name = "com_github_golang_glog",
    importpath = "github.com/golang/glog",
    revision = "23def4e6c14b4da8ac2ed8007337bc5eb5007998",
)

go_internal_tools_deps()

local_repository(
    name = "io_bazel_rules_go",
    path = ".",
)
