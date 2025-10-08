# Описание отношений базы данных

## Таблица user
---
Хранит информацию о пользователях сервиса\
`{id} -> username, name, phone_number, password_hash, description, user_type_id, created_at, updated_at`\
`{username} -> id, name, phone_number, password_hash, description, user_type_id, created_at, updated_at`\
`{phone_number} -> id, username, name, password_hash, description, user_type_id, created_at, updated_at`
- **1НФ** - все атрибуты атомарны, нет составных типов данных
- **2НФ** - все неключевые атрибуты полностью зависят от первичного ключа {id}
- **3НФ** - отсутствуют транзитивные зависимости, все неключевые атрибуты зависят только от ключа
- **НФБК** - все детерминанты (id, username, phone_number) являются потенциальными ключами

## Таблица user_type
---
Хранит типы пользователей (обычный пользователь, пользователь с подпиской)\
`{id} -> name, description, created_at, updated_at`
- **1НФ** - все атрибуты атомарны
- **2НФ** - неключевые атрибуты полностью зависят от первичного ключа
- **3НФ** - нет транзитивных зависимостей
- **НФБК** - отношение находится в 3НФ и имеет один потенциальный ключ

## Таблица chat
---
Хранит информацию о чатах\
`{id} -> chat_type_id, name, description, created_at, updated_at`
- **1НФ** - все атрибуты атомарны
- **2НФ** - все неключевые атрибуты полностью зависят от первичного ключа
- **3НФ** - нет транзитивных зависимостей
- **НФБК** - отношение находится в 3НФ и имеет один потенциальный ключ

## Таблица chat_type
---
Хранит типы чатов (личные сообщения, группа, канал)\
`{id} -> name, description, created_at, updated_at`
- **1НФ** - все атрибуты атомарны
- **2НФ** - неключевые атрибуты полностью зависят от первичного ключа
- **3НФ** - нет транзитивных зависимостей
- **НФБК** - отношение находится в 3НФ и имеет один потенциальный ключ

## Таблица chat_member
---
Хранит информацию об участниках чатов и их ролях\
`{user_id, chat_id} -> chat_member_role_id, created_at, updated_at`
- **1НФ** - все атрибуты атомарны
- **2НФ** - составной первичный ключ, неключевые атрибуты зависят от всего ключа
- **3НФ** - нет транзитивных зависимостей
- **НФБК** - отношение находится в 3НФ и имеет один потенциальный ключ

## Таблица chat_member_role
---
Хранит роли участников чатов (администратор, писатель, наблюдатель)\
`{id} -> name, description, created_at, updated_at`
- **1НФ** - все атрибуты атомарны
- **2НФ** - неключевые атрибуты полностью зависят от первичного ключа
- **3НФ** - нет транзитивных зависимостей
- **НФБК** - отношение находится в 3НФ и имеет один потенциальный ключ

## Таблица message
---
Хранит сообщения в чатах\
`{id} -> chat_id, user_id, text, message_type_id, created_at, updated_at`
- **1НФ** - все атрибуты атомарны
- **2НФ** - все неключевые атрибуты полностью зависят от первичного ключа
- **3НФ** - нет транзитивных зависимостей
- **НФБК** - отношение находится в 3НФ и имеет один потенциальный ключ

## Таблица message_type
---
Хранит типы сообщений (системное или пользовательское)\
`{id} -> name, description, created_at, updated_at`
- **1НФ** - все атрибуты атомарны
- **2НФ** - неключевые атрибуты полностью зависят от первичного ключа
- **3НФ** - нет транзитивных зависимостей
- **НФБК** - отношение находится в 3НФ и имеет один потенциальный ключ

## Таблица session
---
Хранит информацию о сессиях пользователей\
`{id} -> user_id, device, created_at, last_seen`
- **1НФ** - все атрибуты атомарны
- **2НФ** - все неключевые атрибуты полностью зависят от первичного ключа
- **3НФ** - нет транзитивных зависимостей
- **НФБК** - отношение находится в 3НФ и имеет один потенциальный ключ

## Таблица attachment
---
Хранит информацию о файловых вложениях (изображения, документы, аудио, видео)\
`{id} -> file_name, file_size, file_path, content_disposition, created_at, updated_at`
- **1НФ** - все атрибуты атомарны
- **2НФ** - все неключевые атрибуты полностью зависят от первичного ключа
- **3НФ** - нет транзитивных зависимостей
- **НФБК** - отношение находится в 3НФ и имеет один потенциальный ключ

## Таблица avatar_chat
---
Хранит связь между чатами и их аватарами\
`{attachment_id, chat_id} -> created_at, updated_at`
- **1НФ** - все атрибуты атомарны
- **2НФ** - составной первичный ключ, неключевые атрибуты зависят от всего ключа
- **3НФ** - нет транзитивных зависимостей
- **НФБК** - отношение находится в 3НФ и имеет один потенциальный ключ

## Таблица avatar_user
---
Хранит связь между пользователями и их аватарами\
`{attachment_id, user_id} -> created_at, updated_at`
- **1НФ** - все атрибуты атомарны
- **2НФ** - составной первичный ключ, неключевые атрибуты зависят от всего ключа
- **3НФ** - нет транзитивных зависимостей
- **НФБК** - отношение находится в 3НФ и имеет один потенциальный ключ

## Таблица message_attachment
---
Хранит связь между сообщениями и их вложениями\
`{message_id, attachment_id} -> user_id, created_at, updated_at`
- **1НФ** - все атрибуты атомарны
- **2НФ** - составной первичный ключ, неключевые атрибуты зависят от всего ключа
- **3НФ** - нет транзитивных зависимостей
- **НФБК** - отношение находится в 3НФ и имеет один потенциальный ключ

# ER Diagram

```mermaid
erDiagram
    user {
        UUID id PK
        TEXT username UK
        TEXT name
        TEXT phone_number UK
        TEXT password_hash
        TEXT description
        INT4 user_type_id FK
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }
    user_type{
        INT4 id PK
        TEXT name
        TEXT description
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }

    chat {
        UUID id PK
        INT4 chat_type_id FK
        TEXT name
        TEXT description
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }
    chat_type{
        INT4 id PK
        TEXT name
        TEXT description
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }
    
    chat_member{
        UUID user_id FK
        UUID chat_id FK
        INT4 chat_member_role_id FK
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }
    chat_member_role{
        INT4 id PK
        TEXT name
        TEXT description
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }

    message{
        UUID id PK 
        UUID chat_id FK
        UUID user_id FK
        TEXT text
        INT4 message_type_id FK
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }

    message_type{
        INT4 id PK
        TEXT name
        TEXT description
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }

    session {
        UUID id PK
        UUID user_id FK
        TEXT device
        TIMESTAMPTZ created_at
        TIMESTAMPTZ last_seen
    }

    attachment {
        UUID id PK
        TEXT file_name
        BIGINT file_size
        TEXT file_path
        TEXT content_disposition
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }

    avatar_chat {
        UUID attachment_id FK
        UUID chat_id FK
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }

    avatar_user {
        UUID attachment_id FK
        UUID user_id FK
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }

    message_attachment {
        UUID message_id FK
        UUID attachment_id FK
        UUID user_id FK
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }

    user ||--o{ chat_member : "participates_as"
    user ||--o{ message : "author"
    user ||--o{ session : "maintains"
    user ||--o{ avatar_user : "has_avatar"
    chat ||--o{ message : "contains"
    chat ||--o{ avatar_chat : "has_avatar"
    user_type ||--o{ user : "has"
    chat_type ||--o{ chat : "has" 
    chat_member_role ||--o{ chat_member : "has"
    message_type ||--o{ message : "has"
    message ||--o{ message_attachment : "has"
    attachment ||--o{ message_attachment : "attached_to"
    attachment ||--|| avatar_user : "used_as_avatar"
    attachment ||--|| avatar_chat : "used_as_avatar"
```