-- +goose Up
CREATE TABLE departments (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    parent_id BIGINT NULL REFERENCES departments(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX uniq_departments_root_name
    ON departments (lower(name))
    WHERE parent_id IS NULL;

CREATE UNIQUE INDEX uniq_departments_parent_name
    ON departments (parent_id, lower(name))
    WHERE parent_id IS NOT NULL;

CREATE INDEX idx_departments_parent_id ON departments(parent_id);

CREATE TABLE employees (
    id BIGSERIAL PRIMARY KEY,
    department_id BIGINT NOT NULL REFERENCES departments(id) ON DELETE CASCADE,
    full_name VARCHAR(200) NOT NULL,
    position VARCHAR(200) NOT NULL,
    hired_at DATE NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_employees_department_id ON employees(department_id);

-- +goose Down
DROP TABLE IF EXISTS employees;
DROP INDEX IF EXISTS idx_departments_parent_id;
DROP INDEX IF EXISTS uniq_departments_parent_name;
DROP INDEX IF EXISTS uniq_departments_root_name;
DROP TABLE IF EXISTS departments;
