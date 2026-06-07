CREATE TABLE IF NOT EXISTS interviews (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  candidate_id TEXT NOT NULL,
  job_id TEXT NOT NULL,
  questions JSONB NOT NULL DEFAULT '[]'::jsonb,
  answers JSONB NOT NULL DEFAULT '[]'::jsonb,
  technical_score NUMERIC(5,2),
  communication_score NUMERIC(5,2),
  problem_solving_score NUMERIC(5,2),
  overall_score NUMERIC(5,2),
  recommendation VARCHAR(20) NOT NULL DEFAULT '',
  status VARCHAR(20) NOT NULL DEFAULT 'in_progress',
  ai_report JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE interviews ADD COLUMN IF NOT EXISTS candidate_id TEXT;
ALTER TABLE interviews ADD COLUMN IF NOT EXISTS job_id TEXT;
ALTER TABLE interviews ADD COLUMN IF NOT EXISTS questions JSONB NOT NULL DEFAULT '[]'::jsonb;
ALTER TABLE interviews ADD COLUMN IF NOT EXISTS answers JSONB NOT NULL DEFAULT '[]'::jsonb;
ALTER TABLE interviews ADD COLUMN IF NOT EXISTS technical_score NUMERIC(5,2);
ALTER TABLE interviews ADD COLUMN IF NOT EXISTS communication_score NUMERIC(5,2);
ALTER TABLE interviews ADD COLUMN IF NOT EXISTS problem_solving_score NUMERIC(5,2);
ALTER TABLE interviews ADD COLUMN IF NOT EXISTS overall_score NUMERIC(5,2);
ALTER TABLE interviews ADD COLUMN IF NOT EXISTS recommendation VARCHAR(20) NOT NULL DEFAULT '';
ALTER TABLE interviews ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'in_progress';
ALTER TABLE interviews ADD COLUMN IF NOT EXISTS ai_report JSONB NOT NULL DEFAULT '{}'::jsonb;
ALTER TABLE interviews ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE interviews ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

CREATE INDEX IF NOT EXISTS idx_interviews_candidate_id ON interviews (candidate_id);
CREATE INDEX IF NOT EXISTS idx_interviews_job_id ON interviews (job_id);
CREATE INDEX IF NOT EXISTS idx_interviews_status ON interviews (status);
