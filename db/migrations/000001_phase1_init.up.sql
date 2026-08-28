BEGIN;

CREATE TABLE users (
    id UUID PRIMARY KEY,
    wechat_openid VARCHAR(128) NOT NULL UNIQUE,
    status SMALLINT NOT NULL DEFAULT 1
        CHECK (status IN (1, 2)),
    nickname VARCHAR(64),
    avatar_url TEXT,
    gender SMALLINT
        CHECK (gender IN (1, 2, 3)),
    birth_year SMALLINT
        CHECK (birth_year BETWEEN 1900 AND 2100),
    height_cm NUMERIC(5, 2)
        CHECK (height_cm BETWEEN 80 AND 250),
    weight_kg NUMERIC(6, 2)
        CHECK (weight_kg BETWEEN 20 AND 400),
    training_level SMALLINT
        CHECK (training_level IN (1, 2, 3)),
    fitness_goal SMALLINT
        CHECK (fitness_goal IN (1, 2, 3)),
    timezone VARCHAR(64) NOT NULL DEFAULT 'Asia/Shanghai'
        CHECK (btrim(timezone) <> ''),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_login_at TIMESTAMPTZ,
    CHECK (updated_at >= created_at),
    CHECK (last_login_at IS NULL OR last_login_at >= created_at)
);

COMMENT ON COLUMN users.status IS 'UserStatus: 1 active, 2 disabled';
COMMENT ON COLUMN users.gender IS 'Gender: 1 male, 2 female, 3 other; NULL unknown';
COMMENT ON COLUMN users.training_level IS 'TrainingLevel: 1 beginner, 2 intermediate, 3 advanced';
COMMENT ON COLUMN users.fitness_goal IS 'FitnessGoal: 1 fat_loss, 2 muscle_gain, 3 general_fitness';
COMMENT ON COLUMN users.timezone IS 'IANA timezone used to calculate local date ranges';

CREATE TABLE auth_sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash VARCHAR(128) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    device_info JSONB NOT NULL DEFAULT '{}'::jsonb
        CHECK (jsonb_typeof(device_info) = 'object'),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (expires_at > created_at),
    CHECK (revoked_at IS NULL OR revoked_at >= created_at),
    CHECK (last_used_at IS NULL OR last_used_at >= created_at),
    CHECK (updated_at >= created_at)
);

CREATE UNIQUE INDEX uq_auth_sessions_one_unrevoked
    ON auth_sessions(user_id)
    WHERE revoked_at IS NULL;

CREATE INDEX idx_auth_sessions_expires_at
    ON auth_sessions(expires_at);

CREATE TABLE exercises (
    id UUID PRIMARY KEY,
    code VARCHAR(64) NOT NULL UNIQUE
        CHECK (code ~ '^[a-z0-9][a-z0-9_]*$'),
    body_part SMALLINT NOT NULL
        CHECK (body_part BETWEEN 1 AND 9),
    name_zh VARCHAR(100) NOT NULL
        CHECK (btrim(name_zh) <> ''),
    name_en VARCHAR(100) NOT NULL
        CHECK (btrim(name_en) <> ''),
    equipment SMALLINT NOT NULL
        CHECK (equipment IN (1, 2, 3)),
    difficulty SMALLINT NOT NULL
        CHECK (difficulty IN (1, 2)),
    movement_mode SMALLINT NOT NULL
        CHECK (movement_mode IN (1, 2)),
    counting_mode SMALLINT NOT NULL
        CHECK (counting_mode IN (1, 2, 3)),
    default_reps SMALLINT
        CHECK (default_reps BETWEEN 1 AND 999),
    default_hold_seconds INTEGER
        CHECK (default_hold_seconds BETWEEN 1 AND 86400),
    external_load_multiplier SMALLINT NOT NULL DEFAULT 1
        CHECK (external_load_multiplier IN (1, 2)),
    default_phase_plan JSONB NOT NULL DEFAULT '[]'::jsonb
        CHECK (jsonb_typeof(default_phase_plan) = 'array'),
    start_position TEXT NOT NULL,
    steps JSONB NOT NULL DEFAULT '[]'::jsonb
        CHECK (jsonb_typeof(steps) = 'array'),
    breathing_cue TEXT,
    key_cues JSONB NOT NULL DEFAULT '[]'::jsonb
        CHECK (jsonb_typeof(key_cues) = 'array'),
    common_mistakes JSONB NOT NULL DEFAULT '[]'::jsonb
        CHECK (jsonb_typeof(common_mistakes) = 'array'),
    contraindication_note TEXT,
    source_refs JSONB NOT NULL DEFAULT '[]'::jsonb
        CHECK (jsonb_typeof(source_refs) = 'array'),
    review_status SMALLINT NOT NULL DEFAULT 1
        CHECK (review_status IN (1, 2)),
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort_order INTEGER NOT NULL DEFAULT 0
        CHECK (sort_order >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (btrim(start_position) <> ''),
    CHECK (updated_at >= created_at),
    CONSTRAINT ck_exercises_mode_fields CHECK (
        (movement_mode = 1 AND counting_mode IN (1, 2)
            AND default_reps IS NOT NULL AND default_hold_seconds IS NULL)
        OR
        (movement_mode = 2 AND counting_mode = 3
            AND default_reps IS NULL AND default_hold_seconds IS NOT NULL)
    )
);

COMMENT ON COLUMN exercises.body_part IS 'BodyPart: 1 shoulder, 2 back, 3 arm, 4 chest, 5 core, 6 glute, 7 thigh, 8 calf, 9 foot_arch';
COMMENT ON COLUMN exercises.equipment IS 'Equipment: 1 no_equipment, 2 dumbbell, 3 resistance_band';
COMMENT ON COLUMN exercises.difficulty IS 'Difficulty: 1 beginner, 2 intermediate';
COMMENT ON COLUMN exercises.movement_mode IS 'MovementMode: 1 reps, 2 hold';
COMMENT ON COLUMN exercises.counting_mode IS 'CountingMode: 1 total, 2 per_side, 3 seconds';
COMMENT ON COLUMN exercises.external_load_multiplier IS 'Number of same-weight external loads moved in one repetition';
COMMENT ON COLUMN exercises.review_status IS 'ReviewStatus: 1 draft, 2 reviewed';

CREATE INDEX idx_exercises_active_filter
    ON exercises(body_part, equipment, sort_order, id)
    WHERE is_active = true;

CREATE TABLE muscles (
    id UUID PRIMARY KEY,
    code VARCHAR(64) NOT NULL UNIQUE
        CHECK (code ~ '^[a-z0-9][a-z0-9_]*$'),
    name_zh VARCHAR(100) NOT NULL
        CHECK (btrim(name_zh) <> ''),
    name_en VARCHAR(100) NOT NULL
        CHECK (btrim(name_en) <> ''),
    latin_name VARCHAR(150),
    body_region VARCHAR(32) NOT NULL
        CHECK (body_region IN (
            'shoulder', 'upper_back', 'arm', 'chest', 'trunk',
            'hip', 'thigh', 'lower_leg', 'foot'
        )),
    functional_summary TEXT NOT NULL,
    source_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (btrim(functional_summary) <> ''),
    CHECK (btrim(source_url) <> ''),
    CHECK (updated_at >= created_at)
);

CREATE TABLE exercise_muscles (
    exercise_id UUID NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    muscle_id UUID NOT NULL REFERENCES muscles(id) ON DELETE CASCADE,
    role SMALLINT NOT NULL
        CHECK (role IN (1, 2, 3)),
    display_order SMALLINT NOT NULL DEFAULT 0
        CHECK (display_order >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (exercise_id, muscle_id)
);

COMMENT ON COLUMN exercise_muscles.role IS 'MuscleRole: 1 primary, 2 secondary, 3 stabilizer';

CREATE INDEX idx_exercise_muscles_muscle_role
    ON exercise_muscles(muscle_id, role, exercise_id);

CREATE TABLE workout_sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status SMALLINT NOT NULL DEFAULT 1
        CHECK (status IN (1, 2, 3)),
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ,
    discarded_at TIMESTAMPTZ,
    duration_seconds INTEGER
        CHECK (duration_seconds >= 0),
    note TEXT,
    completed_exercise_count INTEGER
        CHECK (completed_exercise_count >= 0),
    completed_set_count INTEGER
        CHECK (completed_set_count >= 0),
    total_reps BIGINT
        CHECK (total_reps >= 0),
    total_volume_kg NUMERIC(16, 2)
        CHECK (total_volume_kg >= 0),
    average_rpe NUMERIC(3, 1)
        CHECK (average_rpe BETWEEN 1 AND 10),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (updated_at >= created_at),
    CHECK (completed_at IS NULL OR completed_at >= started_at),
    CHECK (discarded_at IS NULL OR discarded_at >= started_at),
    CONSTRAINT ck_workout_sessions_state CHECK (
        (status = 1
            AND completed_at IS NULL
            AND discarded_at IS NULL
            AND duration_seconds IS NULL
            AND completed_exercise_count IS NULL
            AND completed_set_count IS NULL
            AND total_reps IS NULL
            AND total_volume_kg IS NULL
            AND average_rpe IS NULL)
        OR
        (status = 2
            AND completed_at IS NOT NULL
            AND discarded_at IS NULL
            AND duration_seconds IS NOT NULL
            AND completed_exercise_count IS NOT NULL
            AND completed_set_count IS NOT NULL
            AND total_reps IS NOT NULL
            AND total_volume_kg IS NOT NULL)
        OR
        (status = 3
            AND completed_at IS NULL
            AND discarded_at IS NOT NULL
            AND duration_seconds IS NULL
            AND completed_exercise_count IS NULL
            AND completed_set_count IS NULL
            AND total_reps IS NULL
            AND total_volume_kg IS NULL
            AND average_rpe IS NULL)
    )
);

COMMENT ON COLUMN workout_sessions.status IS 'WorkoutStatus: 1 active, 2 completed, 3 discarded';
COMMENT ON COLUMN workout_sessions.total_reps IS 'Normalized total reps; per-side reps are multiplied by two';
COMMENT ON COLUMN workout_sessions.total_volume_kg IS 'Frozen external load volume, excluding bodyweight estimation';

CREATE UNIQUE INDEX uq_workout_sessions_one_active
    ON workout_sessions(user_id)
    WHERE status = 1;

CREATE INDEX idx_workout_sessions_history
    ON workout_sessions(user_id, status, completed_at DESC, id);

CREATE TABLE workout_exercises (
    id UUID PRIMARY KEY,
    workout_session_id UUID NOT NULL REFERENCES workout_sessions(id) ON DELETE CASCADE,
    exercise_id UUID NOT NULL REFERENCES exercises(id) ON DELETE RESTRICT,
    exercise_code_snapshot VARCHAR(64) NOT NULL,
    exercise_name_zh_snapshot VARCHAR(100) NOT NULL,
    exercise_name_en_snapshot VARCHAR(100) NOT NULL,
    movement_mode_snapshot SMALLINT NOT NULL
        CHECK (movement_mode_snapshot IN (1, 2)),
    counting_mode_snapshot SMALLINT NOT NULL
        CHECK (counting_mode_snapshot IN (1, 2, 3)),
    external_load_multiplier_snapshot SMALLINT NOT NULL
        CHECK (external_load_multiplier_snapshot IN (1, 2)),
    primary_muscles_snapshot JSONB NOT NULL
        CHECK (
            jsonb_typeof(primary_muscles_snapshot) = 'array'
            AND jsonb_array_length(primary_muscles_snapshot) > 0
        ),
    exercise_order INTEGER NOT NULL
        CHECK (exercise_order >= 1),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (updated_at >= created_at),
    CHECK (
        (movement_mode_snapshot = 1 AND counting_mode_snapshot IN (1, 2))
        OR (movement_mode_snapshot = 2 AND counting_mode_snapshot = 3)
    ),
    UNIQUE (workout_session_id, exercise_id),
    UNIQUE (workout_session_id, exercise_order)
);

CREATE TABLE workout_sets (
    id UUID PRIMARY KEY,
    workout_exercise_id UUID NOT NULL REFERENCES workout_exercises(id) ON DELETE CASCADE,
    set_order INTEGER NOT NULL
        CHECK (set_order >= 1),
    target_reps SMALLINT
        CHECK (target_reps BETWEEN 1 AND 999),
    actual_reps SMALLINT
        CHECK (actual_reps BETWEEN 1 AND 999),
    target_hold_seconds INTEGER
        CHECK (target_hold_seconds BETWEEN 1 AND 86400),
    actual_hold_seconds INTEGER
        CHECK (actual_hold_seconds BETWEEN 1 AND 86400),
    weight_kg NUMERIC(7, 2) NOT NULL DEFAULT 0
        CHECK (weight_kg BETWEEN 0 AND 1000),
    rpe NUMERIC(3, 1)
        CHECK (rpe BETWEEN 1 AND 10),
    status SMALLINT NOT NULL DEFAULT 1
        CHECK (status IN (1, 2, 3, 4)),
    phase_plan_snapshot JSONB NOT NULL DEFAULT '[]'::jsonb
        CHECK (jsonb_typeof(phase_plan_snapshot) = 'array'),
    started_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ,
    actual_duration_seconds INTEGER
        CHECK (actual_duration_seconds >= 0),
    client_request_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (updated_at >= created_at),
    CHECK (ended_at IS NULL OR started_at IS NULL OR ended_at >= started_at),
    CONSTRAINT ck_workout_sets_target_type CHECK (
        (target_reps IS NOT NULL AND target_hold_seconds IS NULL)
        OR (target_reps IS NULL AND target_hold_seconds IS NOT NULL)
    ),
    CONSTRAINT ck_workout_sets_state CHECK (
        (status = 1
            AND started_at IS NULL
            AND ended_at IS NULL
            AND actual_reps IS NULL
            AND actual_hold_seconds IS NULL
            AND actual_duration_seconds IS NULL)
        OR
        (status = 2
            AND started_at IS NOT NULL
            AND ended_at IS NULL
            AND actual_reps IS NULL
            AND actual_hold_seconds IS NULL
            AND actual_duration_seconds IS NULL)
        OR
        (status = 3
            AND ended_at IS NOT NULL
            AND (
                (target_reps IS NOT NULL
                    AND actual_reps IS NOT NULL
                    AND actual_hold_seconds IS NULL)
                OR
                (target_hold_seconds IS NOT NULL
                    AND actual_reps IS NULL
                    AND actual_hold_seconds IS NOT NULL)
            )
            AND (
                (started_at IS NULL AND actual_duration_seconds IS NULL)
                OR
                (started_at IS NOT NULL AND actual_duration_seconds IS NOT NULL)
            ))
        OR
        (status = 4
            AND ended_at IS NOT NULL
            AND actual_reps IS NULL
            AND actual_hold_seconds IS NULL
            AND actual_duration_seconds IS NULL)
    ),
    UNIQUE (workout_exercise_id, set_order),
    UNIQUE (workout_exercise_id, client_request_id)
);

COMMENT ON COLUMN workout_sets.status IS 'WorkoutSetStatus: 1 pending, 2 running, 3 completed, 4 cancelled';
COMMENT ON COLUMN workout_sets.actual_reps IS 'For per_side exercises this value means reps per side';
COMMENT ON COLUMN workout_sets.weight_kg IS 'User-entered single external load weight; 0 for bodyweight or unquantified resistance';

CREATE INDEX idx_workout_sets_status
    ON workout_sets(workout_exercise_id, status, set_order);

COMMIT;
