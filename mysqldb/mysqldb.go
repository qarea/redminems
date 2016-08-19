package mysqldb

import (
	"fmt"

	"../cfg"
	"github.com/jmoiron/sqlx"
)

func New() *sqlx.DB {
	return sqlx.MustConnect("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		cfg.MySQL.Login,
		cfg.MySQL.Password,
		cfg.MySQL.Host,
		cfg.MySQL.Port,
		cfg.MySQL.DB,
	))
}
