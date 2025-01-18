CREATE TABLE torrents (
                          id SERIAL PRIMARY KEY,
                          magnet_link TEXT,
                          info_hash TEXT UNIQUE,
                          status TEXT NOT NULL DEFAULT 'in_progress',
                          created_at TIMESTAMP DEFAULT NOW()
);
