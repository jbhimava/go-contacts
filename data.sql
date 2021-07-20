-- Database tables to store User contact information
-- To store user contact information and minimize the redunduncy userinformation is normalized to two tables as following 

-- Users table to store user information except phone numbers
CREATE TABLE IF NOT EXISTS users (
    id serial PRIMARY KEY,
    first_name varchar(30) NOT NULL,
    last_name varchar(30) NOT NULL,
    email varchar(100) UNIQUE NOT NULL,
    created_at timestamp DEFAULT current_timestamp,
    updated_at timestamp DEFAULT current_timestamp

);

-- normalized user contact information table to store phone numbers for each user
CREATE TABLE IF NOT EXISTS contacts(
    id serial PRIMARY KEY,
    user_id int REFERENCES users (id) ON UPDATE CASCADE ON DELETE CASCADE,
    number varchar(100) NOT NULL,
    created_at timestamp DEFAULT current_timestamp,
    updated_at timestamp DEFAULT current_timestamp
);