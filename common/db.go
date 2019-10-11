package common

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

var (
	UserName = "dgaming"
	Password = "dgaming"
	DBName   = "marketplace"
	DBPort   = 5432
	DBHost   = "postgres"
	//DBHost     = "localhost"
	ConnString = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", DBHost, DBPort, UserName, Password, DBName)
)

func GetDB() (*gorm.DB, error) {
	return gorm.Open("postgres", ConnString)
}
