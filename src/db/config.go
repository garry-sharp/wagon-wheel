package db

import (
	"errors"
	"os"

	"github.com/garry-sharp/assessment/src/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func Init(sqlitefn string) {
	//If DB pointer is null
	if db == nil {

		//If no DB file exists
		if _, err := os.Stat(sqlitefn); errors.Is(err, os.ErrNotExist) {
			os.Create(sqlitefn)
		}

		d, err := gorm.Open(sqlite.Open(sqlitefn), &gorm.Config{})
		if err != nil {
			logger.Log("Cannot connect to DB", logger.Fatal, logger.Price)
		} else {
			//Set up schemas
			d.AutoMigrate(&Price{})
			db = d
		}
	}
}

func GetDB() *gorm.DB {
	return db
}
