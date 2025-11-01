//go:generate mockgen -source=../interface/chats/chats.go -destination=mock_chats_repository.go -package=mocks
//go:generate mockgen -source=../interface/user/user.go -destination=mock_user_repository.go -package=mocks

package mocks
