-- am1 (autonomous materializer, sub-step 1).
--
-- Fire `pg_notify('stewards_pending_file_write', NEW.id::text)` on every
-- INSERT into stewards.pending_file_writes. The bridge LISTENs on this
-- channel and drains the table autonomously (cmd/stewards-mcp). Pairs
-- with a 60s safety poll inside the bridge so a missed NOTIFY (server
-- restart, network blip) doesn't strand the row.
--
-- Idempotent — CREATE OR REPLACE + DROP-and-CREATE-trigger so this file
-- can re-apply cleanly on pg rebuild.
--
-- Companion: cmd/stewards-mcp/bridge_run.go (materializerLoop goroutine).
-- Proposal: .spec/proposals/autonomous-materializer.md (D-AM-3 ratified
-- 2026-05-22).

CREATE OR REPLACE FUNCTION stewards.notify_pending_file_write()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    -- Payload is just the new row id. The bridge ignores it and drains
    -- whatever is pending (since draining is FOR UPDATE SKIP LOCKED, the
    -- exact row is incidental). Keeping the payload non-empty so future
    -- targeted-handling consumers have something to subscribe to.
    PERFORM pg_notify('stewards_pending_file_write', NEW.id::text);
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS pending_file_writes_notify
    ON stewards.pending_file_writes;

CREATE TRIGGER pending_file_writes_notify
    AFTER INSERT ON stewards.pending_file_writes
    FOR EACH ROW
    EXECUTE FUNCTION stewards.notify_pending_file_write();

COMMENT ON FUNCTION stewards.notify_pending_file_write() IS
    'am1 (2026-05-22): fires pg_notify(stewards_pending_file_write) so the bridge can autonomously drain the table. See .spec/proposals/autonomous-materializer.md.';
