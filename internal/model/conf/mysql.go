package conf

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

var DB *gorm.DB

// LogrusWriter 自定义一个实现了 GORM Writer 接口的日志写入器
type LogrusWriter struct {
	logger *logrus.Logger
}

// Printf 实现 GORM Writer 接口的 Printf 方法
func (l *LogrusWriter) Printf(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

func InitMySQL() {
	dsn := viper.GetString("mysql.dsn")
	var err error

	// 创建自定义的 LogrusWriter
	logrusWriter := &LogrusWriter{
		logger: logrus.StandardLogger(),
	}

	newLogger := logger.New(
		logrusWriter,
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info,
			Colorful:      true,
		},
	)

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		logrus.Fatalf("failed to connect database: %v", err)
	}
	// 移除自动迁移代码
	// DB.AutoMigrate(&user_manage.User{})
}
