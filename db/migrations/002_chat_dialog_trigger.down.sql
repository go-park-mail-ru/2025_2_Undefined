-- Удаляем триггеры
DROP TRIGGER IF EXISTS check_dialog_members_trigger ON chat_member;
DROP TRIGGER IF EXISTS check_chat_type_change_trigger ON chat;

-- Удаляем функции
DROP FUNCTION IF EXISTS check_dialog_constraints();
DROP FUNCTION IF EXISTS check_chat_type_change();
DROP FUNCTION IF EXISTS validate_dialog_constraints(UUID);
