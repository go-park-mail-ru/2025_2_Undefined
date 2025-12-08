-- ===========================================
-- Service User Creation Script
-- ===========================================
-- Этот скрипт создает непривилегированного пользователя 'app_user'
-- для работы приложения с базой данных PostgreSQL

-- ===========================================
-- АНАЛИЗ ТРЕБУЕМЫХ ПРАВ
-- ===========================================
--
-- Необходимые операции по таблицам:
--
-- 1. Таблица "user":
--    - SELECT: аутентификация, получение профиля
--    - INSERT: регистрация новых пользователей
--    - UPDATE: обновление профиля (имя, описание, пароль)
--
-- 2. Таблица "chat":
--    - SELECT: получение списка чатов, информации о чате
--    - INSERT: создание новых чатов
--    - UPDATE: изменение названия, описания чата
--
-- 3. Таблица "chat_member":
--    - SELECT: проверка членства, получение списка участников
--    - INSERT: добавление пользователя в чат
--    - UPDATE: изменение роли участника
--    - DELETE: выход из чата, удаление участника
--
-- 4. Таблица "message":
--    - SELECT: получение истории сообщений
--    - INSERT: отправка новых сообщений
--    - UPDATE: редактирование своих сообщений
--    - DELETE: удаление своих сообщений
--
-- 5. Таблица "attachment":
--    - SELECT: получение информации о файлах
--    - INSERT: загрузка новых файлов
--
-- 6. Таблица "avatar_chat":
--    - SELECT: получение аватара чата
--    - INSERT: установка нового аватара
--    - DELETE: удаление старого аватара
--
-- 7. Таблица "avatar_user":
--    - SELECT: получение аватара пользователя
--    - INSERT: установка нового аватара
--    - DELETE: удаление старого аватара
--
-- 8. Таблица "message_attachment":
--    - SELECT: получение файлов, прикрепленных к сообщению
--    - INSERT: прикрепление файла к сообщению
--    - DELETE: открепление файла от сообщения
--
-- 9. Таблица "contact":
--    - SELECT: получение списка контактов
--    - INSERT: добавление в контакты
--    - DELETE: удаление из контактов

-- ===========================================
-- Usage: source .env && psql -U <superuser> -d gramm -v pwd="$POSTGRES_APP_PASSWORD" -f db/create_service_user.sql
-- Создал отдельно скрипт setup_user.sh для удобства
-- ===========================================

-- Разделение ответственности:
-- - Суперпользователь: DDL операции (CREATE/ALTER/DROP), выполнение миграций
-- - app_user: DML операции (SELECT/INSERT/UPDATE/DELETE), используется всеми микросервисами
-- Порядок: сначала миграции создают схему, затем app_user получает права на эти объекты

CREATE USER app_user WITH
    PASSWORD :'pwd'
    NOSUPERUSER
    NOCREATEDB
    NOCREATEROLE
    NOINHERIT
    LOGIN
    CONNECTION LIMIT 490;

COMMENT ON ROLE app_user IS 'Service account - limited DML only';

GRANT USAGE ON SCHEMA public TO app_user;

-- Table "user": auth, profile management
GRANT SELECT, INSERT, UPDATE ON TABLE public."user" TO app_user;

-- Table "chat": list and create chats
GRANT SELECT, INSERT, UPDATE ON TABLE public.chat TO app_user;

-- Table "chat_member": membership management
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.chat_member TO app_user;

-- Table "message": messaging
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.message TO app_user;

-- Table "attachment": file uploads
GRANT SELECT, INSERT ON TABLE public.attachment TO app_user;

-- Table "avatar_chat": chat avatars
GRANT SELECT, INSERT, DELETE ON TABLE public.avatar_chat TO app_user;

-- Table "avatar_user": user avatars
GRANT SELECT, INSERT, DELETE ON TABLE public.avatar_user TO app_user;

-- Table "message_attachment": message files
GRANT SELECT, INSERT, DELETE ON TABLE public.message_attachment TO app_user;

-- Table "contact": contacts list
GRANT SELECT, INSERT, DELETE ON TABLE public.contact TO app_user;

-- Sequences access
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO app_user;
