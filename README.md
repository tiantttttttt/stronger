# Stronger API

居家力量训练记录服务，使用 Go 1.25、Gin、GORM、PostgreSQL 和 JWT。本仓库可独立运行，不依赖前端仓库或外层工作区。

当前实现：开发登录、身份查询、开始空训练、完成训练和读取总结。微信登录、动作与组次记录、历史趋势尚未完成。

模式由 `DEBUG` 控制：`DEBUG=true` 为调试模式，启用 Gin debug 模式和开发登录；`DEBUG=false` 为生产模式，启用 Gin release 模式且不注册开发登录接口。未设置时默认 `false`，其他值会导致启动失败。`.env.example` 使用 `DEBUG=true` 方便本地联调；Compose 同样读取该变量。不再使用 `APP_ENV`，已有 `.env` 请替换为 `DEBUG=true` 或 `DEBUG=false`。

## 目录

```text
main.go        API 入口
consts/        全局数据库连接
pkg/db/        GORM 数据库连接初始化
service/
  api/
    api.go       资源初始化、启动与关闭
    router/      路由注册
    handler/     HTTP 参数与响应
    middleware/  JWT 签发及验证
    logic/       业务逻辑、数据库操作与事务
      requests/  请求结构
      responses/ 响应结构
  model/         实体、状态与业务错误
  config/        环境变量配置
db/           建表 migration 和动作 Seed 数据
scripts/       动作 Seed 校验与 SQL 生成工具
Dockerfile     API 镜像构建
compose.yaml   本地 PostgreSQL 与 API 编排
Makefile       后端开发命令
```

## 本地运行

以下命令均在本服务端仓库根目录执行。

### Docker 运行数据库和 API

需要 Docker Compose v2；动作 Seed 工具需要 Node.js。Go 本地开发需要 Go 1.25，或允许 Go 自动下载对应工具链。

```bash
make env-init
make check
make compose-up
curl http://127.0.0.1:8080/health
```

`make env-init` 仅在 `.env` 不存在时创建文件，并生成 64 个十六进制字符的随机 JWT 密钥；已有文件保持原样。已有文件若使用示例密钥，请用 `openssl rand -hex 32` 生成新密钥，手动替换 `JWT_SECRET`。

Compose 使用 `.env` 读取配置；容器中的数据库地址自动设置为 `db:5432`。API 访问地址为 `http://127.0.0.1:8080`，启动初期可用 `make compose-logs` 查看构建、数据库初始化和启动日志。

数据库首次创建数据卷时自动执行建表 migration。后续启动保留数据，不会重复建表或自动升级表结构。`make compose-down` 停止容器并保留数据卷。此编排仅用于本地开发，开发登录允许指定用户身份，不应对外开放。

如果本机 `5432` 已被占用，可在 `.env` 添加 `POSTGRES_PORT=5433`；容器间连接仍使用 `5432`。

### 本地 Go 调试

只用 Docker 启动数据库，再在本机运行 API：

```bash
make db-up
make api-run
```

数据库完成初始化后才能执行登录和训练操作。确认 `.env` 的 `DATABASE_URL` 指向本机数据库；修改 `POSTGRES_PORT` 后，也要相应修改 URL 中的端口。若已通过 Compose 启动 API，先执行 `make compose-down`，再使用此模式，避免占用同一个 API 端口。

也可以使用自行安装的 PostgreSQL：创建 `stronger` 角色和数据库，配置 `DATABASE_URL`，在空数据库执行以下命令，然后运行 `make api-run`：

```bash
set -a
. ./.env
set +a
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 \
  -f db/migrations/000001_phase1_init.up.sql
```

建表脚本不能重复执行。`/health` 只验证数据库连接，不能代替表结构初始化验证；可继续检查开发登录：

```bash
curl -X POST http://127.0.0.1:8080/api/v1/auth/dev/login \
  -H 'Content-Type: application/json' -d '{}'
```

### 可选动作 Seed

空训练演示不依赖动作 Seed。以下命令先生成 SQL，再导入数据库（需要本机安装 `psql`）：

```bash
set -a
. ./.env
set +a
make seed-validate
make seed-sql > /tmp/stronger-phase1-catalog.sql
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f /tmp/stronger-phase1-catalog.sql
```

Seed 可重复执行，包含 20 个动作、39 个肌肉及关联关系；动作保持草稿、未启用。

## 常用命令

- `make check`：Seed 校验、Go 测试、静态检查和后端构建；需要 Go 与 Node.js。
- `make compose-up`：构建并启动本地数据库和 API；只需要 Docker Compose 和用于初始化密钥的 OpenSSL。
- `make compose-logs`：查看运行日志。
- `make compose-down`：停止容器，保留数据库数据。

API 默认监听 `http://127.0.0.1:8080`。客户端可通过 HTTP 接口独立联调，不需要服务端持有前端源码。

## 数据访问与集成测试

业务代码统一通过 `consts.DB.WithContext(ctx)` 访问 PostgreSQL。`pkg/db.NewDB` 创建连接池，`service/api.New` 在注册路由前初始化全局数据库；单个进程只运行一个服务实例，运行期间不替换连接。关闭时先停止 HTTP 服务，再关闭连接池。

Handler 直接调用 `AuthLogic`、`WorkoutsLogic` 等具体业务单例，不定义业务接口或传递数据库。Logic 使用无状态结构体组织方法，事务和行锁在 `service/api/logic` 中管理；事务内部始终使用局部 `tx`，查询条件、请求上下文和事务不写回全局连接。Auth Handler 保留 Debug 和 TokenManager，Router 通过 `NewRouter(debug, tokens)` 注册路由。

表结构仍由 SQL migration 管理，服务不执行 AutoMigrate。GORM PostgreSQL 驱动内部依赖 pgx，但业务代码不再直接使用 pgx/pgxpool。

现有测试不连接真实数据库，健康检查通过 SQL mock 覆盖连接成功、失败及两秒超时。运行检查：

```bash
make check
go test -race ./...
```

真实数据库验证使用临时脚本和专用 PostgreSQL 测试库，创建独立随机 schema 并应用真实 migration；验证登录、训练闭环、并发、幂等和用户隔离后，只清理测试 schema 并删除临时脚本。临时验证不包含在上述命令中。修改全局数据库的测试串行运行，并在所有请求结束后恢复连接变量。

## 接口文档

Handler 的 swag 注释为文档来源，接口变更后重新生成 `docs/swagger.json` 和 `docs/swagger.yaml`：

```bash
go install github.com/swaggo/swag/cmd/swag@v1.16.4
make swagger
```

确保 Go 的 bin 目录在 PATH 中，或通过 `make swagger SWAG=/实际路径/swag` 指定工具。文档可导入 Swagger UI 或支持 Swagger 2.0 的接口工具；鉴权填写 `Bearer <access_token>`。
