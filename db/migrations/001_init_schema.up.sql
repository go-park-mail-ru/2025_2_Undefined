CREATE TYPE message_type_enum AS ENUM ('user', 'system');
CREATE TYPE chat_member_role_enum AS ENUM ('admin', 'writer', 'viewer');
CREATE TYPE user_type_enum AS ENUM ('user', 'premium', 'verified');
CREATE TYPE chat_type_enum AS ENUM ('channel', 'dialog', 'group');

CREATE TABLE "user" (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    phone_number TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    description TEXT,
    user_type user_type_enum NOT NULL DEFAULT 'user',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT check_username_length CHECK (LENGTH(username) >= 3 AND LENGTH(username) <= 20),
    CONSTRAINT check_name_length CHECK (LENGTH(name) >= 1 AND LENGTH(name) <= 20),
    CONSTRAINT check_phone_format CHECK (phone_number ~ '^\+?[\d\s\-\(\)]{10,20}$'),
    CONSTRAINT check_password_hash_length CHECK (LENGTH(password_hash) = 60), -- bcrypt hash length
    CONSTRAINT check_description_length CHECK (description IS NULL OR LENGTH(description) <= 500)
);

CREATE TABLE chat (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_type chat_type_enum NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT check_chat_name_length CHECK (LENGTH(name) >= 1 AND LENGTH(name) <= 100),
    CONSTRAINT check_chat_description_length CHECK (description IS NULL OR LENGTH(description) <= 1000)
);

CREATE TABLE chat_member (
    user_id UUID NOT NULL REFERENCES "user"(id) ON DELETE CASCADE ON UPDATE CASCADE,
    chat_id UUID NOT NULL REFERENCES chat(id) ON DELETE CASCADE ON UPDATE CASCADE,
    chat_member_role chat_member_role_enum NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, chat_id)
);

CREATE TABLE message (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_id UUID NOT NULL REFERENCES chat(id) ON DELETE CASCADE ON UPDATE CASCADE,
    user_id UUID NOT NULL REFERENCES "user"(id) ON DELETE CASCADE ON UPDATE CASCADE,
    text TEXT NOT NULL,
    message_type message_type_enum NOT NULL DEFAULT 'user',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT check_message_text_length CHECK (LENGTH(text) >= 1 AND LENGTH(text) <= 4000)
);

CREATE TABLE session (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES "user"(id) ON DELETE CASCADE ON UPDATE CASCADE,
    device TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT check_device_length CHECK (LENGTH(device) >= 1 AND LENGTH(device) <= 200)
);

CREATE TABLE attachment (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_name TEXT NOT NULL,
    file_size BIGINT NOT NULL,
    file_path TEXT NOT NULL,
    content_disposition TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT check_file_name_length CHECK (LENGTH(file_name) >= 1 AND LENGTH(file_name) <= 255),
    CONSTRAINT check_file_size_positive CHECK (file_size > 0),
    CONSTRAINT check_file_path_length CHECK (LENGTH(file_path) >= 1 AND LENGTH(file_path) <= 500),
    CONSTRAINT check_content_disposition_length CHECK (LENGTH(content_disposition) >= 1 AND LENGTH(content_disposition) <= 100)
);

CREATE TABLE avatar_chat (
    attachment_id UUID NOT NULL REFERENCES attachment(id) ON DELETE CASCADE ON UPDATE CASCADE,
    chat_id UUID NOT NULL REFERENCES chat(id) ON DELETE CASCADE ON UPDATE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (attachment_id, chat_id)
);

CREATE TABLE avatar_user (
    attachment_id UUID NOT NULL REFERENCES attachment(id) ON DELETE CASCADE ON UPDATE CASCADE,
    user_id UUID NOT NULL REFERENCES "user"(id) ON DELETE CASCADE ON UPDATE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (attachment_id, user_id)
);

CREATE TABLE message_attachment (
    message_id UUID NOT NULL REFERENCES message(id) ON DELETE CASCADE ON UPDATE CASCADE,
    attachment_id UUID NOT NULL REFERENCES attachment(id) ON DELETE CASCADE ON UPDATE CASCADE,
    user_id UUID NOT NULL REFERENCES "user"(id) ON DELETE CASCADE ON UPDATE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (message_id, attachment_id)
);

CREATE TABLE contact (
    user_id UUID NOT NULL REFERENCES "user"(id) ON DELETE CASCADE ON UPDATE CASCADE,
    contact_user_id UUID NOT NULL REFERENCES "user"(id) ON DELETE CASCADE ON UPDATE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, contact_user_id),
    
    CONSTRAINT check_not_self_contact CHECK (user_id != contact_user_id)
);

-- Триггеры
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_user_updated_at BEFORE UPDATE ON "user" FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_chat_updated_at BEFORE UPDATE ON chat FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_message_updated_at BEFORE UPDATE ON message FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_attachment_updated_at BEFORE UPDATE ON attachment FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_avatar_chat_updated_at BEFORE UPDATE ON avatar_chat FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_avatar_user_updated_at BEFORE UPDATE ON avatar_user FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_message_attachment_updated_at BEFORE UPDATE ON message_attachment FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_contact_updated_at BEFORE UPDATE ON contact FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();