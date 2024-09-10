package db

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var globalDB *gorm.DB

func GetDB() (*gorm.DB, error) {
	if globalDB != nil {
		return globalDB, nil
	}

	var err error
	globalDB, err = gorm.Open(sqlite.Open("./db/my-chat.db"))
	if err != nil {
		return nil, fmt.Errorf("open db failed, err = %v", err)
	}

	return globalDB, nil
}
