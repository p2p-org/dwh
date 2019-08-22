package common

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

var (
	UserName   = "dgaming"
	Password   = "dgaming"
	DBName     = "marketplace"
	ConnString = fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", UserName, Password, DBName)
)

func GetDB() (*gorm.DB, error) {
	return gorm.Open("postgres", ConnString)
}
