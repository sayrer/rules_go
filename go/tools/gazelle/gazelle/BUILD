load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_binary", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "diff.go",
        "fix.go",
        "main.go",
        "print.go",
    ],
    deps = [
        "@io_bazel_buildifier//core:go_default_library",
        "@io_bazel_buildifier//differ:go_default_library",
        "//go/tools/gazelle/generator:go_default_library",
    ],
)

go_binary(
    name = "gazelle",
    library = ":go_default_library",
)

go_test(
    name = "gazelle_test",
    srcs = ["fix_test.go"],
    library = ":go_default_library",
)
