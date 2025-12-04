package repository

const (
	getChatsQuery = `
		SELECT c.id, c.chat_type::text, c.name, c.description 
		FROM chat c
		JOIN chat_member cm ON cm.chat_id = c.id
		WHERE cm.user_id = $1`

	getChatQuery = `
		SELECT c.id, c.chat_type::text, c.name, c.description 
		FROM chat c
		WHERE c.id = $1`

	getUsersOfChat = `
		SELECT 
			cm.user_id, cm.chat_id, usr.name, 
			cm.chat_member_role::text
		FROM chat_member cm
		JOIN "user" usr ON usr.id = cm.user_id
		WHERE cm.chat_id = $1`

	getUserInfo = `
		SELECT 
			cm.user_id, cm.chat_id, usr.name, 
			cm.chat_member_role::text
		FROM chat_member cm
		JOIN "user" usr ON usr.id = cm.user_id
		WHERE cm.user_id = $1 AND cm.chat_id = $2`

	getUsersDialogQuery = `
		SELECT chat.id 
		FROM chat
		LEFT JOIN chat_member cm1 ON cm1.chat_id = chat.id
		LEFT JOIN chat_member cm2 ON cm2.chat_id = chat.id
		WHERE cm1.user_id = $1 AND cm2.user_id = $2 AND chat.chat_type = 'dialog'`

	checkUserRoleQuery = `
		SELECT EXISTS(
			SELECT 1 FROM chat_member 
			WHERE user_id = $1 AND chat_id = $2 AND chat_member_role = $3::chat_member_role_enum
		)`

	deleteChatQuery = `DELETE FROM chat WHERE id = $1`

	updateChatQuery = `UPDATE chat SET name = $1, description = $2 WHERE id = $3`

	getChatAvatarsQuery = `
		WITH latest_chat_avatars AS (
			SELECT DISTINCT ON (ac.chat_id) 
				ac.chat_id, 
				a.id as attachment_id
			FROM avatar_chat ac
			JOIN attachment a ON ac.attachment_id = a.id
			WHERE ac.chat_id = ANY($2)
			ORDER BY ac.chat_id, ac.created_at DESC
		),
		latest_user_avatars AS (
			SELECT DISTINCT ON (au.user_id) 
				au.user_id,
				a.id as attachment_id
			FROM avatar_user au
			JOIN attachment a ON au.attachment_id = a.id
			ORDER BY au.user_id, au.created_at DESC
		),
		dialog_avatars AS (
			SELECT 
				c.id as chat_id,
				lua.attachment_id
			FROM chat c
			JOIN chat_member cm ON cm.chat_id = c.id
			JOIN latest_user_avatars lua ON lua.user_id = cm.user_id
			WHERE c.chat_type = 'dialog' 
			  AND c.id = ANY($2)
			  AND cm.user_id != $1
		)
		SELECT chat_id, attachment_id FROM latest_chat_avatars
		WHERE chat_id NOT IN (SELECT chat_id FROM dialog_avatars)
		UNION ALL
		SELECT chat_id, attachment_id FROM dialog_avatars`

	insertChatAvatarInAttachmentTableQuery = `
		INSERT INTO attachment (id, file_name, file_size, content_disposition)
		VALUES ($1, $2, $3, $4)`

	insertChatAvatarInAvatarChatTableQuery = `
		INSERT INTO avatar_chat (chat_id, attachment_id)
		VALUES ($1, $2)`

	searchChatsQuery = `
		SELECT c.id, c.chat_type::text, c.name, c.description 
		FROM chat c
		JOIN chat_member cm ON cm.chat_id = c.id
		WHERE cm.user_id = $1 AND c.name ILIKE '%' || $2 || '%'`
)
