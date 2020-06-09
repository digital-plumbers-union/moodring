################################################################################
# variables
################################################################################

# Project Variables
# Use `env_var` as much as possible to leverage `.env` as single source of truth

# commit := `git rev-parse HEAD`
# branch := `git rev-parse --abbrev-ref HEAD`

################################################################################
# commands
################################################################################

################################################################################
# dependency management recipes
################################################################################

# NOTE: first command is default command
# (i.e., what happens when you run `just` with no recipe)
# update BUILD files & build
build: gazelle
  bazel build //...

test: gazelle
  bazel test //...

gazelle:
  bazel run //:gazelle


# update external go deps in bazel
update-go-deps:
  bazel run //:gazelle -- update-repos -from_file=go.mod -prune=true -to_macro deps.bzl%go --build_file_generation=on --build_file_proto_mode=disable_global

# run basic formatting + linting check against code
check:
  bazel run //:buildifier-check

# run prettier.  THIS WILL UPDATE YO FILES, MAN, SO BE AWARES.
fix:
  bazel run //:buildifier
