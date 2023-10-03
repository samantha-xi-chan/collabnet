package repo_time

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"

	"log"
	"os"
)

var db *gorm.DB
var err error

func Init(dsn string, logLevel int, slowThresholdMs int) {
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold: time.Duration(slowThresholdMs) * time.Millisecond,
				LogLevel:      logger.LogLevel(logLevel), //logger.Info,
			},
		),
	})
	if err != nil {
		log.Println("gorm.Open: ", err)
		panic("gorm.Open, error=" + err.Error())
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(10)

	err = db.Migrator().DropTable(
	//&Time{},
	)
	if err != nil {
		log.Println("gorm.DropTable: ", err)
		panic("DropTable, error=" + err.Error())
	}

	err = db.AutoMigrate(
		&Time{},
	)
	if err != nil {
		log.Println("gorm.AutoMigrate: ", err)
		panic("AutoMigrate, error=" + err.Error())
	}

}

func GetDB(dbs ...*gorm.DB) *gorm.DB {
	if len(dbs) > 0 {
		return dbs[0]
	}
	return db
}
