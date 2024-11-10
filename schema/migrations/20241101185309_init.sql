-- migrate:up transaction:false
CREATE TABLE urls(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    long_url VARCHAR(255) NOT NULL,
    short VARCHAR(7) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME DEFAULT NULL 
);

INSERT INTO
    urls(long_url, short)
VALUES
    ('https://www.google.com', 'abc1234'),
    ('https://www.facebook.com', 'def4567'),
    ('https://www.twitter.com', 'ghi7891');


-- migrate:down

