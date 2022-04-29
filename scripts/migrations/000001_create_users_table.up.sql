CREATE SEQUENCE IF NOT EXISTS user_id_seq;
CREATE TABLE IF NOT EXISTS users (
    id integer NOT NULL DEFAULT nextval('user_id_seq'),
    login text NOT NULL UNIQUE,
    password text NOT NULL,
    PRIMARY KEY (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS login_idx ON users (login);
INSERT INTO users(login, password) VALUES ('cash', '');