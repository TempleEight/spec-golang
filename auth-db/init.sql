CREATE TABLE auth (
  id UUID PRIMARY KEY,
  email TEXT UNIQUE,
  password TEXT
);
