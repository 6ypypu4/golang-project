CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    movie_id INTEGER REFERENCES movies(id) ON DELETE SET NULL,
    review_id INTEGER REFERENCES reviews(id) ON DELETE SET NULL,
    event TEXT NOT NULL,
    details TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_movie_id ON audit_logs(movie_id);


