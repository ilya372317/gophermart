BEGIN;
DROP TRIGGER IF EXISTS set_order_updated_at ON orders;
DROP TABLE IF EXISTS orders;
DROP TYPE IF EXISTS order_status;
COMMIT;