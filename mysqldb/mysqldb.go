package mysqldb

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"gitlab.qarea.org/tgms/planningms/cfg"
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
