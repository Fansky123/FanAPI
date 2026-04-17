-- Add icon_url and description fields to channels table
ALTER TABLE channels ADD COLUMN IF NOT EXISTS icon_url TEXT NOT NULL DEFAULT '';
ALTER TABLE channels ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';
