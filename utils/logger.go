package utils

import (
    "gopkg.in/natefinch/lumberjack.v2"
    "github.com/sirupsen/logrus"
    "os"
    "path/filepath"
)

// NewLogger 初始化并返回一个带有日志滚动的 Logrus 日志记录器
func NewLogger(logFilePath string) *logrus.Logger {
    // 配置 lumberjack 日志滚动器
    logFile := &lumberjack.Logger{
        Filename:   logFilePath,           // 日志文件路径
        MaxSize:    10,                    // 单个日志文件的最大大小（MB）
        MaxBackups: 3,                     // 保留的最大备份数量
        MaxAge:     7,                     // 日志文件的最大保留天数
        Compress:   true,                  // 是否压缩备份日志
    }

    // 创建一个新的 logrus 实例
    logger := logrus.New()

    // 配置 logrus 使用 lumberjack 作为输出
    logger.SetOutput(logFile)

    // 设置日志格式（可选）
    logger.SetFormatter(&logrus.TextFormatter{
        FullTimestamp: true,  // 显示完整时间戳
    })

    // 设置日志级别
    logger.SetLevel(logrus.InfoLevel)

    return logger
}

// GetLogger returns a configured instance of logrus.Logger.
// This logger can be used across the application for consistent logging.
func GetLogger() *logrus.Logger {
    // 获取项目根目录
    cwd, err := os.Getwd()
    if err != nil {
        panic(err)
    }

    // 获取项目根目录的路径
    rootDir := filepath.Dir(cwd)

    // 设置日志文件路径为项目根目录下的 logs/app.log
    logFilePath := filepath.Join(rootDir, "logs", "app.log")

    return NewLogger(logFilePath) // 这里传入日志文件路径
}
