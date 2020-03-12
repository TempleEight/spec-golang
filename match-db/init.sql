CREATE TABLE match (
  id UUID PRIMARY KEY,
  auth_id UUID NOT NULL,
  userOne UUID,
  userTwo UUID,
  matchedOn TIMESTAMP
);
