package message

import (
	"context"
	"log"
	"time"
)

// воркер, который распространяет сообщение всем слушателям чата
func (uc *MessageUsecase) distribute(ctx context.Context) {
	// Fan-out

	log.Println("Distributor started")
	defer log.Println("Distributor stopped")

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-uc.distributeChannel:
			listeners := uc.listenerMap.GetChatListeners(msg.ChatId)
			if listeners == nil {
				continue
			}
			for _, listener := range listeners {
				listener <- msg
			}
		}
	}
}

// воркер, который удаляет пустые чаты
func (uc *MessageUsecase) chatCleaner(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	log.Println("Chat cleaner started")
	defer log.Println("Chat cleaner stopped")

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			uc.listenerMap.CleanInactiveChats()
		}
	}
}

// воркер, который удаляет неактивных пользователей
func (uc *MessageUsecase) readerCleaner(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	log.Println("Reader cleaner started")
	defer log.Println("Reader cleaner stopped")

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			uc.listenerMap.CleanInactiveReaders()
		}
	}
}
