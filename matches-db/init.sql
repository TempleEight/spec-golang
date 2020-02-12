CREATE TABLE Matches (
    id SERIAL PRIMARY KEY,
    userOne int,
    userTwo int,
    matchedOn timestamp
  );

INSERT INTO Matches (userOne, userTwo, matchedOn) VALUES (1, 2, NOW());