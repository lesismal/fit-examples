# 计划

[English](PLAN.md) | [简体中文](PLAN.zh-CN.md)

## 已完成

- **.conf 示例**：每项功能的目录里都有一份 `fit.conf`，由 [examples_test.go](examples_test.go) 用
  `fit validate` 检查。它们从 fit 仓库的 `examples/` 目录迁移而来，Docker Compose 示例的配置也随之从 JSON 改成了 .conf。

## 下一步

- **YAML**：在每个 `fit.conf` 旁边加一份内容相同的 `fit.yaml`。可以先用 `fit adapt -c <目录>/fit.conf -to yaml`
  生成，注释不会保留，需要手工重写；迁移前 fit 的 YAML 示例保存在其历史中，即删除 `examples/` 那次提交之前的
  `examples/*/fit.yaml`。测试随后检查每个 `fit.yaml` 都有效，且转换出的 JSON 与其 `fit.conf` 相同。
- **JSON**：同样在每个目录加一份 `fit.json`（带注释的 JSON），测试相同。
- **运行目录**：目录里有两三个文件后，`fit run -c <目录>` 会把它们合并成一份配置，每个监听器和路由都会出现多次。
  届时 README 要改为说明只运行目录中的一个文件，就像 fit 原来的 README 那样。
