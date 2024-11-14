package utils

import (
    "testing"
    "github.com/sirupsen/logrus"
    "github.com/stretchr/testify/assert"
)

// MockHook 用于捕获日志输出
type MockHook struct {
    Entries []*logrus.Entry
}

// Fire 模拟日志输出
func (hook *MockHook) Fire(entry *logrus.Entry) error {
    hook.Entries = append(hook.Entries, entry)
    return nil
}

// Levels 返回 hook 支持的日志级别
func (hook *MockHook) Levels() []logrus.Level {
    return logrus.AllLevels
}

func TestNewLogger(t *testing.T) {
    // 创建一个 MockHook 实例
    mockHook := &MockHook{}
    
    // 创建一个新的 logger 实例并添加 hook
    logger := NewLogger("test.log")
    logger.AddHook(mockHook)

    // 使用 logger 记录日志
    logger.Info("Test Info Message")
    logger.Warn("Test Warn Message")

    // 验证日志是否被正确记录
    assert.Len(t, mockHook.Entries, 2) // 确保有两个日志条目
    assert.Equal(t, "Test Info Message", mockHook.Entries[0].Message)
    assert.Equal(t, logrus.InfoLevel, mockHook.Entries[0].Level)
    assert.Equal(t, "Test Warn Message", mockHook.Entries[1].Message)
    assert.Equal(t, logrus.WarnLevel, mockHook.Entries[1].Level)
}

func TestGetLogger(t *testing.T) {
    // 使用 GetLogger 获取一个 logger 实例
    logger := GetLogger()

    // 创建一个 MockHook 实例并添加到 logger
    mockHook := &MockHook{}
    logger.AddHook(mockHook)

    // 记录日志
    logger.Info("Info from GetLogger")

    // 验证日志是否被正确记录
    assert.Len(t, mockHook.Entries, 1)
    assert.Equal(t, "Info from GetLogger", mockHook.Entries[0].Message)
    assert.Equal(t, logrus.InfoLevel, mockHook.Entries[0].Level)
}