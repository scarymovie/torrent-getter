CREATE TABLE files (
                       id SERIAL PRIMARY KEY,
                       torrent_id INTEGER REFERENCES torrents(id),
                       name TEXT NOT NULL,
                       size BIGINT NOT NULL,
                       downloaded_size BIGINT DEFAULT 0,
                       status TEXT NOT NULL DEFAULT 'in_progress',
                       path TEXT NOT NULL,
                       UNIQUE (torrent_id, name)
);
