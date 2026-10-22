# 接口设计与实现规范

## 资源命名

资源 Handler 与 Logic 使用相同操作名：Create、Update、Detail、List、Delete；统计查询使用 Stat。按实际能力提供方法，不为凑齐 CRUD 增加接口。认证及基础设施操作保留 DevLogin、Me、Ping 等明确名称。

当前空训练切片提供以下接口，均位于 `/api/v1`：

| HTTP | Handler / Logic | 响应 |
| --- | --- | --- |
| POST /workouts | Workouts.Create | WorkoutRecord |
| PUT /workouts/{id} | Workouts.Update | WorkoutRecord |
| GET /workouts/{id}/stat | Workouts.Stat | WorkoutStat |

Create 保留已有语义：用户行锁内检查档案与 active 场次，已有场次时返回原记录，否则创建空训练。Update 当前只接受 `{"status":2}`，完成时原子冻结空训练汇总，重复完成返回原记录。Stat 仅读取本人已完成场次的冻结数据。完整一期的独立恢复、创建冲突、放弃、列表与详情尚不属于当前切片。

旧 `/workouts/{id}/complete` 和 `/workouts/{id}/summary` 路由已移除，前后端必须同时升级。小程序对应 createWorkout、updateWorkout、getWorkoutStat。写入返回场次身份、状态与时间；需要统计时单独调用 Stat，统计读取失败不应导致自动重放写入。

## 请求与响应

Handler 使用 Gin 的 ShouldBindJSON 绑定请求，Logic 校验字段值、归属和状态。创建场次及开发登录允许省略请求体。

Update 当前仅有 status 可写，数据库更新时间、归属、完成时间和冻结汇总由服务端维护。Handler 的 swag 注释和请求结构体是 Swagger 来源，使用 `make swagger` 生成，不手工修改产物。

WorkoutRecord 与 WorkoutStat 分离。训练响应共用泛型 writeWorkoutResult，训练专属错误映射保留在训练 Handler 中；通用数据库错误不得硬编码训练资源名称。

## swag 注释规范

沿用 coach_server 的精简风格，Handler 仅使用以下六类声明，按示例顺序排列；无鉴权或参数时省略对应行：

```go
// @Summary 更新训练状态
// @Tags workouts
// @Security BearerAuth
// @Param id path string true "训练 UUID" format(uuid)
// @Param body body requests.UpdateWorkout true "更新请求"
// @Success 200 {object} responses.WorkoutRecord
// @Router /api/v1/workouts/{id} [put]
```

- Summary 用一句短语说明用途；参数描述只写简短名称，不展开业务流程。
- 不添加 `@Description` 或逐接口的 `@Failure`。状态转换、幂等及错误语义写在对应规范中，不重复塞入 swag 注释。
- JSON 的 `@accept` / `@produce` 和鉴权定义统一放在 `main.go`，Handler 不重复声明 JSON `@Accept` / `@Produce`。
- 字段必填、枚举等约束维护在请求类型元数据中；路径格式、请求体是否必填、响应类型和路由必须与实现一致。
- 接口修改时只同步受影响的声明，保留已经精简的风格，不主动补回说明性注释。

删除 `@Failure` 仅使 Swagger 不再逐接口列出错误响应，不改变运行时错误处理。修改后在后端目录执行 `make swagger`，同步生成 JSON / YAML，禁止手工编辑生成文件。

## 数据库操作

保持 Router → Handler → Logic → GORM。各资源 Logic 直接使用具体 model 编写查询、创建和更新，不再封装泛型 CRUD 基类。

归属条件、更新字段白名单、状态转换、幂等、行锁和事务均由资源 Logic 控制。事务内部始终使用 tx，不重新取全局连接。

开发用户使用 UpsertDevUser，冲突时仅更新登录时间，不覆盖档案。

## 验证

后端执行 `make check` 和 `go test -race ./...`。当前测试覆盖 UUID 路径参数、健康检查 SQL mock 和配置解析，未保留训练接口契约测试或真实数据库集成测试。

小程序执行 `pnpm run type-check`、`pnpm test`、`pnpm run build:mp-weixin`。客户端测试通过 mock 验证请求路径、方法、字段、鉴权头及错误传播，不代表服务端或数据库验证已通过。
