package main

import (
	"log"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
	_ "github.com/go-park-mail-ru/2025_2_Undefined/docs"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/app"
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
	conf, err := config.NewConfig()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	application, err := app.NewApp(conf)
	if err != nil {
		log.Fatalf("failed to create app: %v", err)
	}

	application.Run()
}
