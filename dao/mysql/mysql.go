package mysql

import (
	"Project1_Shop/settings"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var DB *gorm.DB

func Init(cfg *settings.MySQLConfig) (err error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DbName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名,启用该选项后，`User` 的表名将是 `user`而不是 `users`
		},
	})
	if err != nil {
		return err
	}
	DB = db
	return nil
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
