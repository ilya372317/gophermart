BEGIN;
CREATE TYPE ORDER_STATUS AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');
CREATE TABLE orders
(
    id         BIGSERIAL PRIMARY KEY                              NOT NULL,
    user_id    BIGINT                                             NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE CASCADE,
    status     ORDER_STATUS             DEFAULT ('NEW')           NOT NULL,
    number     BIGINT                                             NOT NULL UNIQUE,
    accrual    DOUBLE PRECISION,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE TRIGGER set_order_updated_at
    BEFORE update
    ON orders
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
COMMIT