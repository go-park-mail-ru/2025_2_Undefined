INSERT INTO "user" (id, username, name, phone_number, password_hash)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'GigaChad',
    'GigaChad',
    '+70000000000',
    '$2a$10$8hlVg4buE55tWxfJE7evUuWkfEoSJBC.pwUedY7ST9D7uI8idPUHe'
);

INSERT INTO appeal_roles (id, user_id, role)
VALUES (
    '00000000-0000-0000-0000-000000000002',
    '00000000-0000-0000-0000-000000000001',
    'admin'
);
