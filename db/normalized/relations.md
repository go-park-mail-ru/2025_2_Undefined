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

    user ||--o{ chat_member : "participates_as"
    user ||--o{ message : "author"
    user ||--o{ session : "maintains"
    chat ||--o{ message : "contains"
    user_type ||--o{ user : "has"
    chat_type ||--o{ chat : "has" 
    chat_member_role ||--o{ chat_member : "has"
    message_type ||--o{ message : "has"

```