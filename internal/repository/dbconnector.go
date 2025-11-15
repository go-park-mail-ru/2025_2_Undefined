package repository

import (
	"fmt"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
)

func GetConnectionString(conf *config.DBConfig) (string, error) {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		conf.User,
		conf.Password,
		conf.Host,
		conf.Port,
		conf.DB,
	), nil
}
