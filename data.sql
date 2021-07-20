
CREATE TABLE IF NOT EXISTS users (
    id serial PRIMARY KEY,
    first_name varchar(30) NOT NULL,
    last_name varchar(30) NOT NULL,
    email varchar(100) UNIQUE NOT NULL,
    created_at timestamp DEFAULT current_timestamp,
    updated_at timestamp DEFAULT current_timestamp

);

CREATE TABLE IF NOT EXISTS contacts(
    id serial PRIMARY KEY,
    user_id int REFERENCES users (id) ON UPDATE CASCADE ON DELETE CASCADE,
    number varchar(100) NOT NULL,
    created_at timestamp DEFAULT current_timestamp,
    updated_at timestamp DEFAULT current_timestamp
);