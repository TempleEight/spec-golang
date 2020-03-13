CREATE TABLE match (
  id UUID PRIMARY KEY,
  created_by UUID NOT NULL,
  userOne UUID,
  userTwo UUID,
  matchedOn TIMESTAMPTZ
);
