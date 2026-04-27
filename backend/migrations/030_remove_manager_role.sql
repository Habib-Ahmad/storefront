-- +goose Up
UPDATE users
SET role = 'staff'
WHERE role = 'manager';

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_role_check;

ALTER TABLE users
    ADD CONSTRAINT users_role_check
    CHECK (role IN ('admin', 'staff'));

-- +goose Down
ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_role_check;

ALTER TABLE users
    ADD CONSTRAINT users_role_check
    CHECK (role IN ('admin', 'staff', 'manager'));