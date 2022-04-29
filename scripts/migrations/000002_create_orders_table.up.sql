CREATE SEQUENCE IF NOT EXISTS order_id_seq;
CREATE TABLE IF NOT EXISTS orders (
    id bigint NOT NULL DEFAULT nextval('order_id_seq'),
    number text NOT NULL UNIQUE,
    user_id integer NOT NULL,
    status text NOT NULL,
    uploaded_at bigint NOT NULL,
    processed_at bigint,
    PRIMARY KEY (id),
    CONSTRAINT fk_user_id
      FOREIGN KEY(user_id)
	      REFERENCES users(id)
);
CREATE UNIQUE INDEX IF NOT EXISTS order_number_idx ON orders (number);