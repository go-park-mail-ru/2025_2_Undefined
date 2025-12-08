-- Удалить pending_attachment таблицу
DROP TABLE IF EXISTS pending_attachment;

ALTER TABLE attachment 
    DROP COLUMN IF EXISTS duration,
    DROP COLUMN IF EXISTS attachment_type;

DROP TYPE IF EXISTS attachment_type_enum;
