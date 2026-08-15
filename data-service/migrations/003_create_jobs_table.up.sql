CREATE TABLE IF NOT EXISTS jobs (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    queue VARCHAR(255) NOT NULL,
    status VARCHAR(255) NOT NULL,
    start_after TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    payload JSONB NOT NULL,
    unique_key TEXT NOT NULL,
    locked_at TIMESTAMP WITH TIME ZONE,
    error TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_jobs_dequeue ON jobs (queue, status, start_after);

CREATE UNIQUE INDEX idx_jobs_unique_key ON jobs (unique_key)
    WHERE status IN ('pending', 'running');

comment on column jobs.id is 'Идентификатор (UUIDv7)';
comment on column jobs.name is 'Название джобы';
comment on column jobs.queue is 'Очередь';
comment on column jobs.status is 'Статус исполнения';
comment on column jobs.start_after is 'Отложенный запуск';
comment on column jobs.payload is 'Json с входными данными';
comment on column jobs.unique_key is 'Ключ дедупликации: пока джоба в очереди или выполняется, вторая такая же не встанет';
comment on column jobs.locked_at is 'Когда джоба взята воркером';
comment on column jobs.error is 'Причина последнего падения';
