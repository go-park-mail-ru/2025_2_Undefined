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

-- Таблицы (в порядке зависимостей: сначала зависимые, потом родительские)
-- Таблицы, которые ссылаются на другие таблицы через внешние ключи
DROP TABLE IF EXISTS message_attachment;  -- ссылается на message, attachment, user
DROP TABLE IF EXISTS avatar_user;        -- ссылается на attachment, user
DROP TABLE IF EXISTS avatar_chat;        -- ссылается на attachment, chat
DROP TABLE IF EXISTS contact;            -- ссылается на user (дважды)
DROP TABLE IF EXISTS message;            -- ссылается на chat, user
DROP TABLE IF EXISTS chat_member;        -- ссылается на user, chat

-- Таблицы без внешних ключей или с минимальными зависимостями
DROP TABLE IF EXISTS attachment;         -- не ссылается на другие таблицы
DROP TABLE IF EXISTS chat;               -- не ссылается на другие таблицы
DROP TABLE IF EXISTS "user";             -- на неё ссылаются другие, удаляем последней

-- Типы перечислений (удаляем после всех таблиц, которые их используют)
DROP TYPE IF EXISTS message_type_enum;
DROP TYPE IF EXISTS chat_member_role_enum;
DROP TYPE IF EXISTS user_type_enum;
DROP TYPE IF EXISTS chat_type_enum;
