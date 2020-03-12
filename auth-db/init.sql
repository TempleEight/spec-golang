CREATE TABLE auth (
  id SERIAL PRIMARY KEY,
  email TEXT UNIQUE,
  password TEXT
);
