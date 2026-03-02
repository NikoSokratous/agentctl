-- Plugin registry and marketplace schema for v0.6

CREATE TABLE IF NOT EXISTS plugins (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    version TEXT NOT NULL,
    author TEXT NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('tool', 'agent', 'integration', 'memory', 'model')),
    metadata JSON NOT NULL,
    artifact_url TEXT,
    downloads INTEGER DEFAULT 0,
    rating REAL DEFAULT 0,
    published_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(name, version)
);

CREATE INDEX IF NOT EXISTS idx_plugins_name ON plugins(name);
CREATE INDEX IF NOT EXISTS idx_plugins_type ON plugins(type);
CREATE INDEX IF NOT EXISTS idx_plugins_author ON plugins(author);
CREATE INDEX IF NOT EXISTS idx_plugins_rating ON plugins(rating DESC);
CREATE INDEX IF NOT EXISTS idx_plugins_downloads ON plugins(downloads DESC);
CREATE INDEX IF NOT EXISTS idx_plugins_published ON plugins(published_at DESC);

CREATE TABLE IF NOT EXISTS plugin_reviews (
    id TEXT PRIMARY KEY,
    plugin_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    rating INTEGER NOT NULL CHECK(rating >= 1 AND rating <= 5),
    comment TEXT,
    helpful INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (plugin_id) REFERENCES plugins(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_plugin_reviews_plugin ON plugin_reviews(plugin_id);
CREATE INDEX IF NOT EXISTS idx_plugin_reviews_user ON plugin_reviews(user_id);
CREATE INDEX IF NOT EXISTS idx_plugin_reviews_rating ON plugin_reviews(rating DESC);
CREATE INDEX IF NOT EXISTS idx_plugin_reviews_created ON plugin_reviews(created_at DESC);

CREATE TABLE IF NOT EXISTS plugin_downloads (
    id TEXT PRIMARY KEY,
    plugin_id TEXT NOT NULL,
    user_id TEXT,
    ip_address TEXT,
    downloaded_at TIMESTAMP NOT NULL,
    FOREIGN KEY (plugin_id) REFERENCES plugins(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_plugin_downloads_plugin ON plugin_downloads(plugin_id);
CREATE INDEX IF NOT EXISTS idx_plugin_downloads_date ON plugin_downloads(downloaded_at DESC);

CREATE TABLE IF NOT EXISTS plugin_tags (
    plugin_id TEXT NOT NULL,
    tag TEXT NOT NULL,
    PRIMARY KEY (plugin_id, tag),
    FOREIGN KEY (plugin_id) REFERENCES plugins(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_plugin_tags_tag ON plugin_tags(tag);
