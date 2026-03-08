package mysql

import (
	"Project1_Shop/models"
	"Project1_Shop/settings"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init(cfg *settings.MySQLConfig) (err error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DbName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	DB = db
	return nil
}

func AutoMigration() {
	// 在这里添加需要自动迁移的模型
	AutoMigrate(
		&models.User{},
		&models.Book{},
		&models.Admin{},
	)
}

func AutoMigrate(models ...interface{}) {
	if err := DB.AutoMigrate(models...); err != nil {
		fmt.Printf("auto migrate failed, err:%v\n", err)
	}
}

func Close() {
	sqlDB, err := DB.DB()
	if err != nil {
		fmt.Printf("get sqlDB failed, err:%v\n", err)
		return
	}
	_ = sqlDB.Close()
}
