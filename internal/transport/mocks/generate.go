//go:generate mockgen -source=../interface/chats/chats.go -destination=mock_chats_usecase.go -package=mocks
//go:generate mockgen -source=../interface/session/session.go -destination=mock_session_usecase.go -package=mocks

package mocks
