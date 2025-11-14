-- Универсальная функция для валидации диалога
CREATE OR REPLACE FUNCTION validate_dialog_constraints(chat_id_param UUID)
RETURNS VOID AS $$
DECLARE
    member_count INTEGER;
    non_admin_count INTEGER;
    existing_dialog_count INTEGER;
    user1_id UUID;
    user2_id UUID;
BEGIN
    -- Подсчитываем количество участников в диалоге
    SELECT COUNT(*) INTO member_count
    FROM chat_member 
    WHERE chat_id = chat_id_param;
    
    -- Подсчитываем количество участников с ролью не 'admin'
    SELECT COUNT(*) INTO non_admin_count
    FROM chat_member 
    WHERE chat_id = chat_id_param 
    AND chat_member_role != 'admin';
    
    -- Проверяем ограничения для диалога
    IF member_count != 2 THEN
        RAISE EXCEPTION 'Dialog must have exactly 2 members, but has %', member_count;
    END IF;
    
    IF non_admin_count > 0 THEN
        RAISE EXCEPTION 'All members in dialog must have admin role, but % members have different roles', non_admin_count;
    END IF;
    
    -- Получаем ID двух участников диалога
    SELECT cm1.user_id, cm2.user_id 
    INTO user1_id, user2_id
    FROM chat_member cm1
    CROSS JOIN chat_member cm2
    WHERE cm1.chat_id = chat_id_param 
    AND cm2.chat_id = chat_id_param
    AND cm1.user_id < cm2.user_id
    LIMIT 1;
    
    -- Проверяем, что между этими пользователями нет других диалогов
    SELECT COUNT(*) INTO existing_dialog_count
    FROM chat c
    INNER JOIN chat_member cm1 ON c.id = cm1.chat_id
    INNER JOIN chat_member cm2 ON c.id = cm2.chat_id
    WHERE c.chat_type = 'dialog'
    AND c.id != chat_id_param
    AND cm1.user_id = user1_id
    AND cm2.user_id = user2_id
    AND (SELECT COUNT(*) FROM chat_member WHERE chat_id = c.id) = 2;
    
    IF existing_dialog_count > 0 THEN
        RAISE EXCEPTION 'Dialog between these users already exists';
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Функция для проверки ограничений диалога при изменении участников
CREATE OR REPLACE FUNCTION check_dialog_constraints()
RETURNS TRIGGER AS $$
DECLARE
    chat_type_value chat_type_enum;
    target_chat_id UUID;
BEGIN
    target_chat_id := COALESCE(NEW.chat_id, OLD.chat_id);
    
    -- Получаем тип чата
    SELECT chat_type INTO chat_type_value 
    FROM chat 
    WHERE id = target_chat_id;
    
    -- Если это не диалог, то проверки не нужны
    IF chat_type_value != 'dialog' THEN
        RETURN COALESCE(NEW, OLD);
    END IF;
    
    -- Дополнительная проверка при INSERT/UPDATE: новый участник должен быть админом
    IF TG_OP IN ('INSERT', 'UPDATE') AND NEW.chat_member_role != 'admin' THEN
        RAISE EXCEPTION 'Dialog member must have admin role, but got %', NEW.chat_member_role;
    END IF;
    
    -- Выполняем общую валидацию диалога
    PERFORM validate_dialog_constraints(target_chat_id);
    
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Триггер на INSERT/UPDATE в таблице chat_member
CREATE TRIGGER check_dialog_members_trigger
    AFTER INSERT OR UPDATE ON chat_member
    FOR EACH ROW
    EXECUTE FUNCTION check_dialog_constraints();

-- Функция для проверки изменения типа чата
CREATE OR REPLACE FUNCTION check_chat_type_change()
RETURNS TRIGGER AS $$
BEGIN
    -- Если тип чата изменяется на 'dialog', проверяем ограничения
    IF NEW.chat_type = 'dialog' AND (OLD.chat_type IS NULL OR OLD.chat_type != 'dialog') THEN
        -- Используем общую функцию валидации
        PERFORM validate_dialog_constraints(NEW.id);
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Триггер на UPDATE в таблице chat
CREATE TRIGGER check_chat_type_change_trigger
    BEFORE UPDATE ON chat
    FOR EACH ROW
    EXECUTE FUNCTION check_chat_type_change();
