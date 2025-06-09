CREATE TABLE
  news (
    id BIGSERIAL PRIMARY KEY,
    content text NOT NULL,
    category text,
    create_at timestamp NOT NULL
  );