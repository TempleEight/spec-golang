CREATE TABLE match (
  id SERIAL PRIMARY KEY,
  auth_id INT NOT NULL,
  userOne INT,
  userTwo INT,
  matchedOn TIMESTAMP,
);
