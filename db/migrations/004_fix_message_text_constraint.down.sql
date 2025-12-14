-- Возвращаем старое ограничение: текст должен быть от 1 до 4000 символов
ALTER TABLE message DROP CONSTRAINT IF EXISTS check_message_text_length;

ALTER TABLE message ADD CONSTRAINT check_message_text_length 
    CHECK (LENGTH(text) >= 1 AND LENGTH(text) <= 4000);

ALTER TABLE attachment DROP CONSTRAINT IF EXISTS check_file_size_positive;

ALTER TABLE attachment ADD CONSTRAINT check_file_size_positive 
    CHECK (file_size > 0);