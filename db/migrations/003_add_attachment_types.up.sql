-- Создание ENUM для типов вложений
CREATE TYPE attachment_type_enum AS ENUM (
    'image',
    'document',
    'audio',
    'video',
    'sticker',
    'voice',
    'video_note'
);

ALTER TABLE attachment 
    ADD COLUMN attachment_type attachment_type_enum NULL,
    ADD COLUMN duration INTEGER;

ALTER TABLE attachment
    ADD CONSTRAINT check_duration_positive CHECK (duration IS NULL OR duration > 0);

-- Таблица для вложений, которые загружены, но еще не привязаны к сообщению
CREATE TABLE pending_attachment (
    attachment_id UUID PRIMARY KEY REFERENCES attachment(id) ON DELETE CASCADE ON UPDATE CASCADE,
    user_id UUID NOT NULL REFERENCES "user"(id) ON DELETE CASCADE ON UPDATE CASCADE,
    created_at6 TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Индекс для поиска старых неиспользованных вложений
CREATE INDEX idx_pending_attachment_created_at ON pending_attachment(created_at);

-- Триггер для автообновления updated_at
CREATE TRIGGER update_pending_attachment_updated_at 
    BEFORE UPDATE ON pending_attachment 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE pending_attachment IS 'Вложения, загруженные через POST /messages/attachment, но еще не привязанные к сообщению';
COMMENT ON COLUMN attachment.attachment_type IS 'Тип вложения: image, document, audio, video, sticker, voice, video_note';
COMMENT ON COLUMN attachment.duration IS 'Длительность в секундах для audio, voice, video_note';
