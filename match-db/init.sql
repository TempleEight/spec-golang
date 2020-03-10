CREATE TABLE match (
    id SERIAL PRIMARY KEY,
    auth_id INT NOT NULL,
    userOne int,
    userTwo int,
    matchedOn timestamp
  );
