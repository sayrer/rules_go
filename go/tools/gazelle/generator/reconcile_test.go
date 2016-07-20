package generator

import (
	"testing"

	bzl "github.com/bazelbuild/buildifier/core"
)

func TestReconcileLoad(t *testing.T) {
	var g Generator
	for _, spec := range []struct {
		dest, generated, want string
	}{
		{
			dest:      ``,
			generated: `go_library(name = "foo")`,
			want:      `load("@io_bazel_rules_go//go:def.bzl", "go_library")`,
		},
		{
			dest: `load("@io_bazel_rules_go//go:def.bzl", "go_prefix")`,
			generated: `
				go_prefix("example.com/foo")
				go_library(name = "foo")
			`,
			want: `load("@io_bazel_rules_go//go:def.bzl", "go_prefix", "go_library")`,
		},
		{
			dest: `
				package(default_visibility = ["//visibility:public"])

				cc_library(
					name = "foo",
					srcs = ["foo.cc"],
				)
			`,
			generated: `go_prefix("example.com/foo")`,
			want: `
				package(default_visibility = ["//visibility:public"])

				load("@io_bazel_rules_go//go:def.bzl", "go_prefix")

				cc_library(
					name = "foo",
					srcs = ["foo.cc"],
				)
			`,
		},
	} {
		dest, err := bzl.Parse("BUILD", []byte(spec.dest))
		if err != nil {
			t.Fatalf("bzl.Parse(%q, %q) failed with %v; want success", "BUILD", spec.dest, err)
		}
		gen, err := bzl.Parse("BUILD", []byte(spec.generated))
		if err != nil {
			t.Fatalf("bzl.Parse(%q, %q) failed with %v; want success", "BUILD", spec.generated, err)
		}
		w, err := bzl.Parse("BUILD", []byte(spec.want))
		if err != nil {
			t.Fatalf("bzl.Parse(%q, %q) failed with %v; want success", "BUILD", spec.want, err)
		}

		g.reconcileLoad(dest, gen)
		if got, want := string(bzl.Format(dest)), string(bzl.Format(w)); got != want {
			t.Errorf("d = %s; want %s; spec = %#v", got, want, spec)
		}
	}
}

func TestReconcileRules(t *testing.T) {
	var g Generator
	for _, spec := range []struct {
		dest, generated, want string
	}{
		{
			dest: `
				cc_library(
					name = "a",
					srcs = ["a.cc"],
				)
			`,
			generated: `
				go_library(
					name = "foo",
					srcs = ["foo.go"],
				)
			`,
			want: `
				cc_library(
					name = "a",
					srcs = ["a.cc"],
				)

				go_library(
					name = "foo",
					srcs = ["foo.go"],
				)
			`,
		},
		{
			dest: `
				cc_library(
					name = "a",
					srcs = ["a.cc"],
				)

				go_library(
					name = "foo",
					srcs = ["foo.go"],
				)

				cc_library(
					name = "b",
					srcs = ["b.cc"],
				)
			`,
			generated: `
				go_library(
					name = "foo",
					srcs = ["foo.go", "bar.go"],
				)
			`,
			want: `
				cc_library(
					name = "a",
					srcs = ["a.cc"],
				)

				go_library(
					name = "foo",
					srcs = ["foo.go", "bar.go"],
				)

				cc_library(
					name = "b",
					srcs = ["b.cc"],
				)
			`,
		},
		{
			dest: `
				go_library(
					name = "foo",
					srcs = ["foo.go", "bar.go"],  # some note
					deps = [":baz"],
					visibility = ["//example:__package__"],
					licenses = ["reciprocal"],  # MIT
				)
			`,
			generated: `
				go_library(
					name = "foo",
					srcs = ["foo.go"],
				)
			`,
			want: `
				go_library(
					name = "foo",
					srcs = ["foo.go"],  # some note
					visibility = ["//example:__package__"],
					licenses = ["reciprocal"],  # MIT
				)
			`,
		},
	} {
		dest, err := bzl.Parse("BUILD", []byte(spec.dest))
		if err != nil {
			t.Fatalf("bzl.Parse(%q, %q) failed with %v; want success", "BUILD", spec.dest, err)
		}
		gen, err := bzl.Parse("BUILD", []byte(spec.generated))
		if err != nil {
			t.Fatalf("bzl.Parse(%q, %q) failed with %v; want success", "BUILD", spec.generated, err)
		}
		w, err := bzl.Parse("BUILD", []byte(spec.want))
		if err != nil {
			t.Fatalf("bzl.Parse(%q, %q) failed with %v; want success", "BUILD", spec.want, err)
		}

		g.reconcileRules(dest, gen)
		if got, want := string(bzl.Format(dest)), string(bzl.Format(w)); got != want {
			t.Errorf("d = %s; want %s; spec = %#v", got, want, spec)
		}
	}
}
