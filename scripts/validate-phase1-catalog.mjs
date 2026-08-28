import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';

const catalogPath = fileURLToPath(new URL('../db/seeds/phase1_catalog.plan.json', import.meta.url));
const catalog = JSON.parse(readFileSync(catalogPath, 'utf8'));
const codePattern = /^[a-z0-9][a-z0-9_]*$/;

function fail(message) {
  throw new Error(message);
}

function uniqueCodes(items, label) {
  const seen = new Set();
  for (const item of items) {
    if (!codePattern.test(item.code)) fail(`${label} code 非法: ${item.code}`);
    if (seen.has(item.code)) fail(`${label} code 重复: ${item.code}`);
    seen.add(item.code);
  }
  return seen;
}

function nonEmptyString(value, label) {
  if (typeof value !== 'string' || value.trim() === '') fail(`${label} 必须是非空字符串`);
}

function stringArray(value, label, { nonEmpty = false } = {}) {
  if (!Array.isArray(value)) fail(`${label} 必须是数组`);
  if (nonEmpty && value.length === 0) fail(`${label} 不能为空`);
  value.forEach((item, index) => nonEmptyString(item, `${label}[${index}]`));
}

if (catalog.version !== 1) fail('catalog version 必须为 1');
if (catalog.status !== 'draft') fail('评审阶段 catalog status 必须为 draft');
if (catalog.review_status !== 1) fail('评审阶段 review_status 必须为 1 draft');
if (catalog.is_active !== false) fail('未经审核的 catalog 不得启用');
if (!Array.isArray(catalog.muscles) || catalog.muscles.length === 0) fail('muscles 不能为空');
if (!Array.isArray(catalog.exercises) || catalog.exercises.length !== 20) fail('一期必须恰好包含 20 个动作');

const muscleCodes = uniqueCodes(catalog.muscles, 'muscle');
uniqueCodes(catalog.exercises, 'exercise');

for (const muscle of catalog.muscles) {
  nonEmptyString(muscle.name_zh, `${muscle.code}.name_zh`);
}

for (const exercise of catalog.exercises) {
  nonEmptyString(exercise.name_zh, `${exercise.code}.name_zh`);
  if (!Number.isInteger(exercise.body_part) || exercise.body_part < 1 || exercise.body_part > 9) fail(`${exercise.code}: body_part 非法`);
  if (!Number.isInteger(exercise.equipment) || ![1, 2, 3].includes(exercise.equipment)) fail(`${exercise.code}: equipment 非法`);
  if (!Number.isInteger(exercise.difficulty) || ![1, 2].includes(exercise.difficulty)) fail(`${exercise.code}: difficulty 非法`);
  if (!Number.isInteger(exercise.external_load_multiplier) || ![1, 2].includes(exercise.external_load_multiplier)) fail(`${exercise.code}: external_load_multiplier 非法`);

  const isReps = exercise.movement_mode === 1;
  const isHold = exercise.movement_mode === 2;
  if (!isReps && !isHold) fail(`${exercise.code}: movement_mode 非法`);
  if (isReps && ![1, 2].includes(exercise.counting_mode)) fail(`${exercise.code}: reps 动作 counting_mode 必须为 total 或 per_side`);
  if (isHold && exercise.counting_mode !== 3) fail(`${exercise.code}: hold 动作 counting_mode 必须为 seconds`);
  if (isReps && (!Number.isInteger(exercise.default_reps) || exercise.default_reps < 1 || exercise.default_hold_seconds !== undefined)) fail(`${exercise.code}: reps 默认值非法`);
  if (isHold && (!Number.isInteger(exercise.default_hold_seconds) || exercise.default_hold_seconds < 1 || exercise.default_reps !== undefined)) fail(`${exercise.code}: hold 默认值非法`);

  stringArray(exercise.primary, `${exercise.code}.primary`, { nonEmpty: true });
  stringArray(exercise.supporting ?? [], `${exercise.code}.supporting`);
  const relationships = [...exercise.primary, ...(exercise.supporting ?? [])];
  if (new Set(relationships).size !== relationships.length) fail(`${exercise.code}: 肌肉关系重复`);
  for (const muscleCode of relationships) {
    if (!muscleCodes.has(muscleCode)) fail(`${exercise.code}: 未声明的 muscle code ${muscleCode}`);
  }
}

console.log(`phase1 catalog valid: ${catalog.exercises.length} exercises, ${catalog.muscles.length} muscles, ${catalog.status}, inactive`);
