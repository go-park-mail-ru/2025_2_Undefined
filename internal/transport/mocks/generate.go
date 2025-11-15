//go:generate mockgen -source=../interface/chats/chats.go -destination=mock_chats_usecase.go -package=mocks
//go:generate mockgen -source=../interface/session/session.go -destination=mock_session_usecase.go -package=mocks
//go:generate mockgen -source=../interface/user/user.go -destination=mock_user_usecase.go -package=mocks
//go:generate mockgen -source=../interface/message/message.go -destination=mock_message_usecase.go -package=mocks

package mocks
