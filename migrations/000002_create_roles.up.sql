CREATE TABLE roles (
    id   serial PRIMARY KEY,
    name text NOT NULL UNIQUE
);

CREATE TABLE permissions (
    id   serial PRIMARY KEY,
    name text NOT NULL UNIQUE
);

CREATE TABLE user_roles (
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id int  NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE role_permissions (
    role_id       int NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id int NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

INSERT INTO roles (name) VALUES ('user'), ('admin');