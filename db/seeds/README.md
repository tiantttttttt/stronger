# 第一期动作目录 Seed

## 文件定位

- `phase1_catalog.plan.json`：首批 20 个动作及其肌肉关系的结构化评审清单。
- `scripts/render-phase1-catalog-seed.mjs`：根据评审清单生成带固定 UUID、可重复执行的 PostgreSQL seed。
- `scripts/validate-phase1-catalog.mjs`：不依赖第三方包的结构校验命令。

当前 JSON 是 seed 的内容规划源，不直接写入数据库。生成器会补充一期草案元数据，所有动作保持 `draft + inactive`；完成动作说明、来源和专业复核后，才允许改为 `reviewed + active`。

## 稳定性规则

- 动作和肌肉 `code` 发布后不得改变含义或重复利用。
- 正式 seed 使用固定 UUID；重复执行不得创建重复记录。
- 每个动作至少包含一个 primary 肌肉。
- reps 动作只能使用 total 或 per_side；hold 动作只能使用 seconds。
- 哑铃重量统一表示单只重量，`external_load_multiplier` 表示一次重复同时移动的同重量哑铃数量。
- 内容初始状态为 draft；生产启用前必须完成专业审核和来源检查。
- 已被训练历史引用的动作只能停用，不能物理删除。

## 校验

```bash
node scripts/validate-phase1-catalog.mjs
```

校验通过只代表结构和引用完整，不代表动作说明已经完成专业审核。

生成或导入 seed：

```bash
node scripts/render-phase1-catalog-seed.mjs > /tmp/phase1_catalog.sql
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f /tmp/phase1_catalog.sql
```

相同版本可以重复执行；动作和肌肉按稳定 `code` 更新，关系以当前规划文件为准重建。生成器不会在重复执行时覆盖数据库中已经完成的 `review_status` 和 `is_active` 发布状态。
