CREATE TABLE user_temple (
  id UUID PRIMARY KEY,
  name TEXT
);

CREATE TABLE picture (
  id UUID PRIMARY KEY,
  user_id UUID REFERENCES user_temple(id),
  img BYTEA
);
