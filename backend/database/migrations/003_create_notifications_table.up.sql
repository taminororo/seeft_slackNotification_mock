CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    shift_id INTEGER NOT NULL REFERENCES shifts(id) ON DELETE CASCADE,
    year_id INTEGER NOT NULL,
    time_id INTEGER NOT NULL,
    date VARCHAR(50) NOT NULL,
    weather VARCHAR(50) NOT NULL,
    old_task_name VARCHAR(255),
    new_task_name VARCHAR(255) NOT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    slack_dm_sent BOOLEAN DEFAULT FALSE,
    slack_channel_sent BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_shift_id ON notifications(shift_id);
CREATE INDEX idx_notifications_is_read ON notifications(is_read);
CREATE INDEX idx_notifications_user_is_read ON notifications(user_id, is_read);

