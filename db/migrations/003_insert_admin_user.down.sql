-- Rollback for 003_insert_admin_user.up.sql
-- Removes the inserted appeal_roles entry and the user

DELETE FROM appeal_roles
WHERE user_id = '00000000-0000-0000-0000-000000000001';

DELETE FROM "user"
WHERE id = '00000000-0000-0000-0000-000000000001';
