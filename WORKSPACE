workspace(
    name = "moodring",
)

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive", "http_file")

################################################################################
# BAZEL RULE / TOOLCHAIN SETUP
################################################################################

# download `io_bazel_rules_go` up front to ensure all of our other rulesets
# leverage the same version, see related issue:
# https://github.com/bazelbuild/rules_go/issues/2398#issuecomment-597139571
http_archive(
    name = "io_bazel_rules_go",
    sha256 = "7b9bbe3ea1fccb46dcfa6c3f3e29ba7ec740d8733370e21cdc8937467b4a4349",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.22.4/rules_go-v0.22.4.tar.gz",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.22.4/rules_go-v0.22.4.tar.gz",
    ],
)

#########################################
# DOCKER
#########################################

http_archive(
    name = "io_bazel_rules_docker",
    sha256 = "dc97fccceacd4c6be14e800b2a00693d5e8d07f69ee187babfd04a80a9f8e250",
    strip_prefix = "rules_docker-0.14.1",
    urls = ["https://github.com/bazelbuild/rules_docker/releases/download/v0.14.1/rules_docker-v0.14.1.tar.gz"],
)

load(
    "@io_bazel_rules_docker//repositories:repositories.bzl",
    container_repositories = "repositories",
)

# configures the docker toolchain, https://github.com/nlopezgi/rules_docker/blob/master/toolchains/docker/readme.md#how-to-use-the-docker-toolchain
container_repositories()

# This is NOT needed when going through the language lang_image
# "repositories" function(s).
load("@io_bazel_rules_docker//repositories:deps.bzl", container_deps = "deps")

container_deps()

load(
    "@io_bazel_rules_docker//container:container.bzl",
    "container_pull",
)

#########################################
# GOLANG
#########################################

# set up `io_bazel_rules_go` imported at top of this file
load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")

# gazell generates BUILD files for go/protobuf
http_archive(
    name = "bazel_gazelle",
    urls = [
        "https://storage.googleapis.com/bazel-mirror/github.com/bazelbuild/bazel-gazelle/releases/download/0.20.0/bazel-gazelle-0.20.0.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/0.20.0/bazel-gazelle-0.20.0.tar.gz",
    ],
)

# protobuf
http_archive(
    name = "com_google_protobuf",
    sha256 = "b0a1da830747a2ffc1125fc84dbd3fe32a876396592d4580501749a2d0d0cb15",
    strip_prefix = "protobuf-3.12.2",
    urls = ["https://github.com/protocolbuffers/protobuf/archive/v3.12.2.zip"],
)

load("@com_google_protobuf//:protobuf_deps.bzl", "protobuf_deps")
load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")

# only set up dependencies once we have imported everything that could
# possibly be overridden: see "overriding dependencies" here:
# https://github.com/bazelbuild/rules_go/blob/master/go/workspace.rst#id9
go_rules_dependencies()

go_register_toolchains()

gazelle_dependencies()

protobuf_deps()

# pull buildtools now that all dependencies are loaded/installed:
# - go
# - gazelle
# - protobuf
http_archive(
    name = "com_github_bazelbuild_buildtools",
    strip_prefix = "buildtools-master",
    url = "https://github.com/bazelbuild/buildtools/archive/master.zip",
)

# enables packaging of various formats (tar, etc)
http_archive(
    name = "rules_pkg",
    sha256 = "352c090cc3d3f9a6b4e676cf42a6047c16824959b438895a76c2989c6d7c246a",
    urls = [
        "https://github.com/bazelbuild/rules_pkg/releases/download/0.2.5/rules_pkg-0.2.5.tar.gz",
        "https://mirror.bazel.build/github.com/bazelbuild/rules_pkg/releases/download/0.2.5/rules_pkg-0.2.5.tar.gz",
    ],
)

load("@rules_pkg//:deps.bzl", "rules_pkg_dependencies")

rules_pkg_dependencies()

#########################################
# SKYLIB - UTILITIES / FUNCTIONS FOR WRITING .BZL FILES
#########################################

load("@bazel_skylib//:workspace.bzl", "bazel_skylib_workspace")

bazel_skylib_workspace()

################################################################################
# EXTERNAL DEPENDENCIES
################################################################################

load(":deps.bzl", "go")

##########################################################
# GO DEPENDENCIES
##########################################################

# this function is generated using gazelle, we load and execute it here to
# reduce WORKSPACE file size
go()

# pull in go_image deps
# see: https://github.com/bazelbuild/rules_docker#go_image

load(
    "@io_bazel_rules_docker//go:image.bzl",
    _go_image_repos = "repositories",
)

_go_image_repos()
