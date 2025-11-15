package main

import (
	"context"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
	_ "github.com/go-park-mail-ru/2025_2_Undefined/docs"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/app"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
)

// @title           Undefined team API documentation of project Telegram
// @version         1.0
// @description     API сервер для чат-приложения в стиле Telegram. Позволяет регистрировать пользователей, управлять чатами и обмениваться сообщениями.
// @termsOfService  http://swagger.io/terms/

// @contact.name   Undefined Team
// @contact.url    https://github.com/go-park-mail-ru/2025_2_Undefined

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1
func main() {
	const op = "main"
	ctx := context.Background()
	logger := domains.GetLogger(ctx).WithField("operation", op)

	conf, err := config.NewConfig()
	if err != nil {
		logger.WithError(err).Fatal("config error")
	}

	application, err := app.NewApp(conf)
	if err != nil {
		logger.WithError(err).Fatal("failed to create app")
	}

	application.Run()
}
