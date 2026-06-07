-- ─── Run this in pgAdmin ──────────────────────────────────────────────────────

-- Add status column to users if it doesn't exist
ALTER TABLE users ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'active';

-- Attendance table
CREATE TABLE IF NOT EXISTS attendance (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    check_in    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    check_out   TIMESTAMPTZ,
    date        DATE        NOT NULL DEFAULT CURRENT_DATE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (employee_id, date)
);

-- Leaves table
CREATE TABLE IF NOT EXISTS leaves (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    leave_type  VARCHAR(20) NOT NULL CHECK (leave_type IN ('sick', 'casual', 'earned')),
    start_date  DATE        NOT NULL,
    end_date    DATE        NOT NULL,
    reason      TEXT        NOT NULL,
    status      VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    approved_by UUID        REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Jobs table (needed for dashboard count — populate later)
CREATE TABLE IF NOT EXISTS jobs (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    title      VARCHAR(255) NOT NULL,
    status     VARCHAR(20)  NOT NULL DEFAULT 'open',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS department VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS required_skills TEXT NOT NULL DEFAULT '';
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS location VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS posted_by UUID REFERENCES users(id);

-- Candidates table (needed for dashboard count — populate later)
CREATE TABLE IF NOT EXISTS candidates (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(255) NOT NULL,
    email      VARCHAR(255) NOT NULL,
    job_id     UUID         REFERENCES jobs(id),
    status     VARCHAR(20)  NOT NULL DEFAULT 'applied',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
ALTER TABLE candidates ADD COLUMN IF NOT EXISTS phone VARCHAR(30) NOT NULL DEFAULT '';
ALTER TABLE candidates ADD COLUMN IF NOT EXISTS resume_url TEXT NOT NULL DEFAULT '';
ALTER TABLE candidates ADD COLUMN IF NOT EXISTS resume_text TEXT NOT NULL DEFAULT '';
ALTER TABLE candidates ADD COLUMN IF NOT EXISTS ai_score NUMERIC(5, 2);
ALTER TABLE candidates ADD COLUMN IF NOT EXISTS ai_report JSONB NOT NULL DEFAULT '{}'::jsonb;
ALTER TABLE candidates ADD COLUMN IF NOT EXISTS cover_letter TEXT NOT NULL DEFAULT '';

-- Payroll table
CREATE TABLE IF NOT EXISTS payroll (
    id           UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id  UUID           NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    month        INT            NOT NULL CHECK (month BETWEEN 1 AND 12),
    year         INT            NOT NULL CHECK (year >= 2000),
    basic_salary NUMERIC(12, 2) NOT NULL CHECK (basic_salary > 0),
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    UNIQUE (employee_id, month, year)
);

-- Notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title      VARCHAR(255) NOT NULL,
    message    TEXT         NOT NULL,
    is_read    BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_notifications_user_read
    ON notifications(user_id, is_read, created_at DESC);

-- Performance reviews table
CREATE TABLE IF NOT EXISTS performance_reviews (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id   UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reviewer_id   UUID        NOT NULL REFERENCES users(id),
    review_period VARCHAR(50) NOT NULL,
    rating        INT         NOT NULL CHECK (rating BETWEEN 1 AND 5),
    feedback      TEXT        NOT NULL,
    goals         TEXT        NOT NULL DEFAULT '',
    status        VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'submitted', 'approved')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
ALTER TABLE performance_reviews ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'draft';
ALTER TABLE performance_reviews ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

CREATE INDEX IF NOT EXISTS idx_performance_reviews_employee_id
    ON performance_reviews(employee_id);
