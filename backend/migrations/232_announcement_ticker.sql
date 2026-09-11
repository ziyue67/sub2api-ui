-- Configure whether an announcement appears in the user header ticker.
-- Existing announcements remain eligible by default; administrators can opt
-- individual announcements out or raise priority to pin them near the front.
ALTER TABLE announcements
    ADD COLUMN IF NOT EXISTS ticker_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS priority INTEGER NOT NULL DEFAULT 0;

ALTER TABLE announcements
    DROP CONSTRAINT IF EXISTS announcements_priority_range;
ALTER TABLE announcements
    ADD CONSTRAINT announcements_priority_range CHECK (priority >= 0 AND priority <= 100);

CREATE INDEX IF NOT EXISTS announcements_ticker_order_idx
    ON announcements (ticker_enabled, priority DESC, id DESC);
