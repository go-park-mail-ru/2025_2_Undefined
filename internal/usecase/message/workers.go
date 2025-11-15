package message

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
)

// воркер, который распространяет сообщение всем слушателям чата
func (uc *MessageUsecase) distribute(ctx context.Context) {
	const op = "MessageUsecase.distribute"

	// Fan-out
	logger := domains.GetLogger(ctx).WithField("operation", op)

	logger.Info("Distributor started")
	defer logger.Info("Distributor stopped")

	for {
		select {
		case <-ctx.Done():
			return
		case websocketMsg := <-uc.distributeChannel:
			logger.Info("distribute message to listeners")

			listeners := uc.listenerMap.GetChatListeners(websocketMsg.ChatID)
			logger.Debugf("listeners %v of message %v", websocketMsg, listeners)
			if listeners == nil {
				continue
			}

			for _, listener := range listeners {
				logger.Debugf("send message %v to listener %v", websocketMsg, listener)
				listener <- websocketMsg
			}
		}
	}
}

// воркер, который удаляет пустые чаты
func (uc *MessageUsecase) chatCleaner(ctx context.Context) {
	const op = "MessageUsecase.chatCleaner"

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	logger := domains.GetLogger(ctx).WithField("operation", op)

	logger.Info("Chat cleaner started")
	defer logger.Infof("Chat cleaner stopped")

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
	const op = "MessageUsecase.readerCleaner"

	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	logger := domains.GetLogger(ctx).WithField("operation", op)

	logger.Info("Chat cleaner started")
	defer logger.Infof("Chat cleaner stopped")

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			logger.Infof("Active disctributers to out channels (goroutines): %d", uc.distributersToOutChannelsCount.Load())
			uc.listenerMap.CleanInactiveReaders()
		}
	}
}
