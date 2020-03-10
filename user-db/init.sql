CREATE TABLE user_temple_auth (
    id SERIAL PRIMARY KEY, 
    email TEXT UNIQUE,
    password TEXT
);

CREATE TABLE user_temple (
    id INT REFERENCES user_temple_auth(id),
    name TEXT
);
