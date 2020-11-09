CREATE TABLE devices
(
    id integer CONSTRAINT id_pk PRIMARY KEY,
    name character varying CONSTRAINT name_nn NOT NULL,
    token character varying CONSTRAINT token_nn NOT NULL,
    created_at timestamp,
    online bool DEFAULT false
);