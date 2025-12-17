-- Удаляем старое ограничение на длину текста
ALTER TABLE message DROP CONSTRAINT IF EXISTS check_message_text_length;

-- Добавляем новое ограничение: текст может быть пустым, но не более 4000 символов
-- Для сообщений со стикерами, голосовыми и видео-кружками текст может быть пустой строкой
ALTER TABLE message ADD CONSTRAINT check_message_text_length 
    CHECK (LENGTH(text) >= 0 AND LENGTH(text) <= 4000);
    
ALTER TABLE attachment DROP CONSTRAINT IF EXISTS check_file_size_positive;

ALTER TABLE attachment ADD CONSTRAINT check_file_size_positive
    CHECK (file_size >= 0);