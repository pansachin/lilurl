-- migrate:up
UPDATE sqlite_sequence SET seq = 1000000 WHERE name = 'urls';


-- migrate:down

