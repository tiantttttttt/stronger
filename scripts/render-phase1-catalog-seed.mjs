import { createHash } from 'node:crypto';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';

const catalogPath = fileURLToPath(new URL('../db/seeds/phase1_catalog.plan.json', import.meta.url));
const catalog = JSON.parse(readFileSync(catalogPath, 'utf8'));

const sourceByRegion = {
  shoulder: 'https://openstax.org/books/anatomy-and-physiology-2e/pages/11-5-muscles-of-the-pectoral-girdle-and-upper-limbs',
  upper_back: 'https://openstax.org/books/anatomy-and-physiology-2e/pages/11-3-axial-muscles-of-the-head-neck-and-back',
  arm: 'https://openstax.org/books/anatomy-and-physiology-2e/pages/11-5-muscles-of-the-pectoral-girdle-and-upper-limbs',
  chest: 'https://openstax.org/books/anatomy-and-physiology-2e/pages/11-4-axial-muscles-of-the-abdominal-wall-and-thorax',
  trunk: 'https://openstax.org/books/anatomy-and-physiology-2e/pages/11-4-axial-muscles-of-the-abdominal-wall-and-thorax',
  hip: 'https://openstax.org/books/anatomy-and-physiology-2e/pages/11-6-appendicular-muscles-of-the-pelvic-girdle-and-lower-limbs',
  thigh: 'https://openstax.org/books/anatomy-and-physiology-2e/pages/11-6-appendicular-muscles-of-the-pelvic-girdle-and-lower-limbs',
  lower_leg: 'https://openstax.org/books/anatomy-and-physiology-2e/pages/11-6-appendicular-muscles-of-the-pelvic-girdle-and-lower-limbs',
  foot: 'https://pmc.ncbi.nlm.nih.gov/articles/PMC5094843/',
};

const muscleMetadata = {
  deltoid_anterior: ['Anterior deltoid', 'shoulder'],
  deltoid_middle: ['Middle deltoid', 'shoulder'],
  deltoid_posterior: ['Posterior deltoid', 'shoulder'],
  supraspinatus: ['Supraspinatus', 'shoulder'],
  trapezius_upper: ['Upper trapezius', 'upper_back'],
  latissimus_dorsi: ['Latissimus dorsi', 'upper_back'],
  rhomboids: ['Rhomboids', 'upper_back'],
  trapezius_middle: ['Middle trapezius', 'upper_back'],
  trapezius_lower: ['Lower trapezius', 'upper_back'],
  erector_spinae: ['Erector spinae', 'trunk'],
  biceps_brachii: ['Biceps brachii', 'arm'],
  brachialis: ['Brachialis', 'arm'],
  brachioradialis: ['Brachioradialis', 'arm'],
  triceps_brachii: ['Triceps brachii', 'arm'],
  pectoralis_major_sternal: ['Sternal head of pectoralis major', 'chest'],
  pectoralis_major_clavicular: ['Clavicular head of pectoralis major', 'chest'],
  serratus_anterior: ['Serratus anterior', 'chest'],
  rectus_abdominis: ['Rectus abdominis', 'trunk'],
  transversus_abdominis: ['Transversus abdominis', 'trunk'],
  internal_oblique: ['Internal oblique', 'trunk'],
  external_oblique: ['External oblique', 'trunk'],
  multifidus: ['Multifidus', 'trunk'],
  gluteus_maximus: ['Gluteus maximus', 'hip'],
  gluteus_medius: ['Gluteus medius', 'hip'],
  gluteus_minimus: ['Gluteus minimus', 'hip'],
  deep_hip_external_rotators: ['Deep hip external rotators', 'hip'],
  quadriceps: ['Quadriceps femoris', 'thigh'],
  biceps_femoris: ['Biceps femoris', 'thigh'],
  semitendinosus: ['Semitendinosus', 'thigh'],
  semimembranosus: ['Semimembranosus', 'thigh'],
  adductor_magnus: ['Adductor magnus', 'thigh'],
  gastrocnemius: ['Gastrocnemius', 'lower_leg'],
  soleus: ['Soleus', 'lower_leg'],
  tibialis_anterior: ['Tibialis anterior', 'lower_leg'],
  abductor_hallucis: ['Abductor hallucis', 'foot'],
  abductor_digiti_minimi: ['Abductor digiti minimi', 'foot'],
  flexor_digitorum_brevis: ['Flexor digitorum brevis', 'foot'],
  flexor_hallucis_brevis: ['Flexor hallucis brevis', 'foot'],
  lumbricals_foot: ['Foot lumbricals', 'foot'],
};

const exerciseNames = {
  dumbbell_lateral_raise: 'Dumbbell lateral raise',
  pike_push_up: 'Pike push-up',
  one_arm_dumbbell_row: 'One-arm dumbbell row',
  prone_reverse_fly: 'Prone reverse fly',
  dumbbell_curl: 'Dumbbell curl',
  close_grip_push_up: 'Close-grip push-up',
  standard_push_up: 'Standard push-up',
  dumbbell_floor_press: 'Dumbbell floor press',
  forearm_plank: 'Forearm plank',
  dead_bug: 'Dead bug',
  bird_dog: 'Bird dog',
  glute_bridge: 'Glute bridge',
  side_lying_clamshell: 'Side-lying clamshell',
  bodyweight_squat: 'Bodyweight squat',
  reverse_lunge: 'Reverse lunge',
  dumbbell_romanian_deadlift: 'Dumbbell Romanian deadlift',
  standing_calf_raise: 'Standing calf raise',
  bent_knee_calf_raise: 'Bent-knee calf raise',
  short_foot: 'Short foot exercise',
  toe_yoga: 'Toe yoga',
};

function stableUuid(kind, code) {
  const bytes = createHash('sha256').update(`stronger:phase1:${kind}:${code}`).digest().subarray(0, 16);
  bytes[6] = (bytes[6] & 0x0f) | 0x50;
  bytes[8] = (bytes[8] & 0x3f) | 0x80;
  const hex = bytes.toString('hex');
  return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`;
}

function sqlString(value) {
  if (value === null || value === undefined) return 'NULL';
  return `'${String(value).replaceAll("'", "''")}'`;
}

function json(value) {
  return `${sqlString(JSON.stringify(value))}::jsonb`;
}

function values(rows) {
  return rows.map((row) => `  (${row.join(', ')})`).join(',\n');
}

const muscleRows = catalog.muscles.map((muscle) => {
  const metadata = muscleMetadata[muscle.code];
  if (!metadata) throw new Error(`缺少 muscle metadata: ${muscle.code}`);
  const [nameEn, region] = metadata;
  return [
    sqlString(stableUuid('muscle', muscle.code)),
    sqlString(muscle.code),
    sqlString(muscle.name_zh),
    sqlString(nameEn),
    'NULL',
    sqlString(region),
    sqlString(`${muscle.name_zh}的一期功能说明草案，生产启用前必须完成专业审核。`),
    sqlString(sourceByRegion[region]),
  ];
});

const exerciseRows = catalog.exercises.map((exercise, index) => {
  const nameEn = exerciseNames[exercise.code];
  if (!nameEn) throw new Error(`缺少 exercise name_en: ${exercise.code}`);
  const phasePlan = exercise.movement_mode === 1
    ? [{ cue: '发力', seconds: 2 }, { cue: '还原', seconds: 2 }]
    : [];
  return [
    sqlString(stableUuid('exercise', exercise.code)),
    sqlString(exercise.code),
    exercise.body_part,
    sqlString(exercise.name_zh),
    sqlString(nameEn),
    exercise.equipment,
    exercise.difficulty,
    exercise.movement_mode,
    exercise.counting_mode,
    exercise.default_reps ?? 'NULL',
    exercise.default_hold_seconds ?? 'NULL',
    exercise.external_load_multiplier,
    json(phasePlan),
    sqlString(`${exercise.name_zh}的起始姿势草案，生产启用前必须完成专业审核。`),
    json([`${exercise.name_zh}的执行步骤草案，保持动作可控并在不适时停止。`]),
    sqlString('保持自然呼吸，避免憋气；生产启用前需专业复核。'),
    json(['保持动作可控', '使用无痛活动范围']),
    json(['避免借力过大', '避免在疼痛状态下继续']),
    sqlString('存在疼痛、损伤或医疗限制时，应先咨询合格专业人员。'),
    json(['https://www.acefitness.org/resources/everyone/exercise-library/', 'https://exrx.net/Lists/Directory']),
    catalog.review_status,
    String(catalog.is_active),
    (index + 1) * 10,
  ];
});

const relationshipRows = [];
for (const exercise of catalog.exercises) {
  exercise.primary.forEach((muscleCode, index) => relationshipRows.push([
    sqlString(exercise.code), sqlString(muscleCode), 1, index + 1,
  ]));
  (exercise.supporting ?? []).forEach((muscleCode, index) => relationshipRows.push([
    sqlString(exercise.code), sqlString(muscleCode), 2, index + 1,
  ]));
}

process.stdout.write(`-- Generated from server/db/seeds/phase1_catalog.plan.json.
-- Draft content remains inactive until task 4.3 professional review passes.
BEGIN;

INSERT INTO muscles (
  id, code, name_zh, name_en, latin_name, body_region, functional_summary, source_url
) VALUES
${values(muscleRows)}
ON CONFLICT (code) DO UPDATE SET
  name_zh = EXCLUDED.name_zh,
  name_en = EXCLUDED.name_en,
  latin_name = EXCLUDED.latin_name,
  body_region = EXCLUDED.body_region,
  functional_summary = EXCLUDED.functional_summary,
  source_url = EXCLUDED.source_url,
  updated_at = now();

INSERT INTO exercises (
  id, code, body_part, name_zh, name_en, equipment, difficulty,
  movement_mode, counting_mode, default_reps, default_hold_seconds,
  external_load_multiplier, default_phase_plan, start_position, steps,
  breathing_cue, key_cues, common_mistakes, contraindication_note,
  source_refs, review_status, is_active, sort_order
) VALUES
${values(exerciseRows)}
ON CONFLICT (code) DO UPDATE SET
  body_part = EXCLUDED.body_part,
  name_zh = EXCLUDED.name_zh,
  name_en = EXCLUDED.name_en,
  equipment = EXCLUDED.equipment,
  difficulty = EXCLUDED.difficulty,
  movement_mode = EXCLUDED.movement_mode,
  counting_mode = EXCLUDED.counting_mode,
  default_reps = EXCLUDED.default_reps,
  default_hold_seconds = EXCLUDED.default_hold_seconds,
  external_load_multiplier = EXCLUDED.external_load_multiplier,
  default_phase_plan = EXCLUDED.default_phase_plan,
  start_position = EXCLUDED.start_position,
  steps = EXCLUDED.steps,
  breathing_cue = EXCLUDED.breathing_cue,
  key_cues = EXCLUDED.key_cues,
  common_mistakes = EXCLUDED.common_mistakes,
  contraindication_note = EXCLUDED.contraindication_note,
  source_refs = EXCLUDED.source_refs,
  sort_order = EXCLUDED.sort_order,
  updated_at = now();

DELETE FROM exercise_muscles em
USING exercises e
WHERE em.exercise_id = e.id
  AND e.code IN (${catalog.exercises.map((exercise) => sqlString(exercise.code)).join(', ')});

INSERT INTO exercise_muscles (exercise_id, muscle_id, role, display_order)
SELECT e.id, m.id, seed.role, seed.display_order
FROM (VALUES
${values(relationshipRows)}
) AS seed(exercise_code, muscle_code, role, display_order)
JOIN exercises e ON e.code = seed.exercise_code
JOIN muscles m ON m.code = seed.muscle_code
ON CONFLICT (exercise_id, muscle_id) DO UPDATE SET
  role = EXCLUDED.role,
  display_order = EXCLUDED.display_order;

COMMIT;
`);
