BEGIN;
CREATE TABLE withdrawals
(
    id           BIGSERIAL PRIMARY KEY    NOT NULL,
    order_number BIGINT                   NOT NULL,
    sum          BIGINT                   NOT NULL,
    user_id      BIGINT                   NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE CASCADE,
    created_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TRIGGER set_withdrawals_updated_at
    BEFORE UPDATE
    ON withdrawals
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
COMMIT;