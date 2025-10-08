-- Вставляем типы пользователей
INSERT INTO user_type (name, description) VALUES
('regular', 'Regular user'),
('premium', 'User with subscription');

-- Вставляем типы чатов
INSERT INTO chat_type (name, description) VALUES
('channel', 'Broadcast channel'),
('private', 'Private chat between two users'),
('group', 'Group chat with multiple users');

-- Вставляем роли участников чата
INSERT INTO chat_member_role (name, description) VALUES
('admin', 'Administrator with full permissions'),
('writer', 'Can write messages'),
('reader', 'Can only read messages');

-- Вставляем типы сообщений
INSERT INTO message_type (name, description) VALUES
('user', 'User generated message'),
('system', 'System generated message');