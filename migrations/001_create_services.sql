CREATE TABLE IF NOT EXISTS monitored_service (
  name TEXT PRIMARY KEY,
  endpoint TEXT NOT NULL,
  frequency_seconds INTEGER NOT NULL,
  emails TEXT[],
  last_status TEXT,
  last_checked TIMESTAMP
);
