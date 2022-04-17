CREATE SEQUENCE IF NOT EXISTS balance_journal_id_seq;
CREATE TABLE IF NOT EXISTS balance_journal (
    id bigint NOT NULL DEFAULT nextval('balance_journal_id_seq'),
    type text NOT NULL,
    PRIMARY KEY (id)
);

CREATE SEQUENCE IF NOT EXISTS posting_id_seq;
CREATE TABLE IF NOT EXISTS posting (
    id bigint NOT NULL DEFAULT nextval('posting_id_seq'),
    user_id integer NOT NULL,
    journal_id bigint NOT NULL,
    amount money NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_journal_id
      FOREIGN KEY(journal_id)
	    REFERENCES balance_journal(id),
    CONSTRAINT fk_user_id
      FOREIGN KEY(user_id)
	    REFERENCES users(id)
);