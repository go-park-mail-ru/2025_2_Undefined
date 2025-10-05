CREATE TABLE user_type (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE chat_type (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE chat_member_role (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE message_type (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE "user" (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    phone_number TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    description TEXT,
    user_type_id UUID NOT NULL REFERENCES user_type(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE chat (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_type_id UUID NOT NULL REFERENCES chat_type(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE chat_member (
    user_id UUID NOT NULL REFERENCES "user"(id) ON DELETE CASCADE ON UPDATE CASCADE,
    chat_id UUID NOT NULL REFERENCES chat(id) ON DELETE CASCADE ON UPDATE CASCADE,
    chat_member_role_id UUID NOT NULL REFERENCES chat_member_role(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, chat_id)
);

CREATE TABLE message (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_id UUID NOT NULL REFERENCES chat(id) ON DELETE CASCADE ON UPDATE CASCADE,
    user_id UUID NOT NULL REFERENCES "user"(id) ON DELETE CASCADE ON UPDATE CASCADE,
    text TEXT NOT NULL,
    message_type_id UUID NOT NULL REFERENCES message_type(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE session (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES "user"(id) ON DELETE CASCADE ON UPDATE CASCADE,
    device TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Индексы
CREATE INDEX idx_user_username ON "user"(username);
CREATE INDEX idx_user_phone_number ON "user"(phone_number);
CREATE INDEX idx_chat_member_user_id ON chat_member(user_id);
CREATE INDEX idx_chat_member_chat_id ON chat_member(chat_id);
CREATE INDEX idx_message_chat_id ON message(chat_id);
CREATE INDEX idx_message_user_id ON message(user_id);
CREATE INDEX idx_message_created_at ON message(created_at);
CREATE INDEX idx_session_user_id ON session(user_id);
CREATE INDEX idx_session_last_seen ON session(last_seen);

-- Триггеры
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_user_updated_at BEFORE UPDATE ON "user" FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_user_type_updated_at BEFORE UPDATE ON user_type FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_chat_updated_at BEFORE UPDATE ON chat FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_chat_type_updated_at BEFORE UPDATE ON chat_type FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_chat_member_updated_at BEFORE UPDATE ON chat_member FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_chat_member_role_updated_at BEFORE UPDATE ON chat_member_role FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_message_updated_at BEFORE UPDATE ON message FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_message_type_updated_at BEFORE UPDATE ON message_type FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();