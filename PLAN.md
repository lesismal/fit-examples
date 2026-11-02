# Plan

[English](PLAN.md) | [简体中文](PLAN.zh-CN.md)

## Done

- **.conf examples**: every feature's directory has its config as `fit.conf`, checked with
  `fit validate` by [examples_test.go](examples_test.go). They moved here from fit's `examples/`
  directory, the Docker Compose example's config from JSON to .conf with them.

## Next

- **YAML**: a `fit.yaml` beside every `fit.conf`, saying the same. `fit adapt -c <dir>/fit.conf
  -to yaml` gives a start, without the comments, which are written again by hand; the YAML
  examples fit had before the move are in its history, at `examples/*/fit.yaml` before the commit
  that removed `examples/`. The test then checks that each `fit.yaml` validates and adapts to the
  same JSON as its `fit.conf`.
- **JSON**: a `fit.json` (JSON with comments) beside each, in the same way, with the same test.
- **Running a directory**: once a directory has two or three files, `fit run -c <dir>` joins them
  into one config, with every listener and route given more than once. The README then says to
  run one file of a directory, as fit's did.
