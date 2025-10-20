-- Сначала триггеры
DROP TRIGGER IF EXISTS update_user_updated_at ON "user";
DROP TRIGGER IF EXISTS update_chat_updated_at ON chat;
DROP TRIGGER IF EXISTS update_message_updated_at ON message;
DROP TRIGGER IF EXISTS update_attachment_updated_at ON attachment;
DROP TRIGGER IF EXISTS update_avatar_chat_updated_at ON avatar_chat;
DROP TRIGGER IF EXISTS update_avatar_user_updated_at ON avatar_user;
DROP TRIGGER IF EXISTS update_message_attachment_updated_at ON message_attachment;
DROP TRIGGER IF EXISTS update_contact_updated_at ON contact;

-- Удаляем функции
DROP FUNCTION IF EXISTS update_updated_at_column;

-- Таблицы
DROP TABLE IF EXISTS message_attachment;
DROP TABLE IF EXISTS avatar_user;
DROP TABLE IF EXISTS avatar_chat;
DROP TABLE IF EXISTS attachment;
DROP TABLE IF EXISTS session;
DROP TABLE IF EXISTS message;
DROP TABLE IF EXISTS chat_member;
DROP TABLE IF EXISTS chat;
DROP TABLE IF EXISTS "user";
DROP TABLE IF EXISTS contact;

-- Типы перечислений
DROP TYPE IF EXISTS chat_type_enum;
DROP TYPE IF EXISTS user_type_enum;
DROP TYPE IF EXISTS chat_member_role_enum;
DROP TYPE IF EXISTS message_type_enum;
