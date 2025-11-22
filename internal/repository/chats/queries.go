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
		WHERE cm1.user_id = $1 AND cm2.user_id = $2`

	checkUserRoleQuery = `
		SELECT EXISTS(
			SELECT 1 FROM chat_member 
			WHERE user_id = $1 AND chat_id = $2 AND chat_member_role = $3::chat_member_role_enum
		)`

	deleteChatQuery = `DELETE FROM chat WHERE id = $1`

	updateChatQuery = `UPDATE chat SET name = $1, description = $2 WHERE id = $3`

	getChatAvatarsQuery = `
		WITH latest_avatars AS (
			SELECT DISTINCT ON (ac.chat_id) 
				ac.chat_id, 
				a.id as attachment_id
			FROM avatar_chat ac
			JOIN attachment a ON ac.attachment_id = a.id
			WHERE ac.chat_id = ANY($1)
			ORDER BY ac.chat_id, ac.created_at DESC
		)
		SELECT chat_id, attachment_id 
		FROM latest_avatars`

	insertChatAvatarInAttachmentTableQuery = `
		INSERT INTO attachment (id, file_name, file_size, content_disposition)
		VALUES ($1, $2, $3, $4)`

	insertChatAvatarInAvatarChatTableQuery = `
		INSERT INTO avatar_chat (chat_id, attachment_id)
		VALUES ($1, $2)`
)
