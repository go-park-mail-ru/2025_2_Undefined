//go:generate mockgen -source=../interface/chats/chats.go -destination=mock_chats_repository.go -package=mocks
//go:generate mockgen -source=../interface/user/user.go -destination=mock_user_repository.go -package=mocks
//go:generate mockgen -source=../interface/message/message.go -destination=mock_message_repository.go -package=mocks
//go:generate mockgen -source=../interface/listener/listener.go -destination=mock_listener_map.go -package=mocks
//go:generate mockgen -source=../interface/storage/storage.go -destination=mock_storage.go -package=mocks
//go:generate mockgen -source=../interface/contact/contact.go -destination=mock_contact_repository.go -package=mocks

package mocks
