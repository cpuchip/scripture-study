-- +goose Up
CREATE TABLE audit_log (
    id BIGSERIAL PRIMARY KEY,
    table_name TEXT NOT NULL,
    row_id INTEGER NOT NULL,
    operation TEXT NOT NULL CHECK (operation IN ('UPDATE', 'DELETE')),
    old_data JSONB NOT NULL,
    changed_by INTEGER,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_log_table_row ON audit_log(table_name, row_id);
CREATE INDEX idx_audit_log_changed_at ON audit_log(changed_at);

-- Audit function: captures OLD row as JSON before UPDATE/DELETE
CREATE OR REPLACE FUNCTION audit_trigger_func() RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO audit_log (table_name, row_id, operation, old_data, changed_by)
    VALUES (
        TG_TABLE_NAME,
        OLD.id,
        TG_OP,
        row_to_json(OLD)::jsonb,
        NULLIF(current_setting('app.current_user_id', true), '')::integer
    );
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Attach to practices and tasks tables
CREATE TRIGGER practices_audit
    BEFORE UPDATE OR DELETE ON practices
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_func();

CREATE TRIGGER tasks_audit
    BEFORE UPDATE OR DELETE ON tasks
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_func();

-- +goose Down
DROP TRIGGER IF EXISTS practices_audit ON practices;
DROP TRIGGER IF EXISTS tasks_audit ON tasks;
DROP FUNCTION IF EXISTS audit_trigger_func();
DROP TABLE IF EXISTS audit_log;
