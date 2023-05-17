package log

import (
	"fmt"
	"github.com/ThreeKing2018/gocolor"
	"io"
	"log"
	"os"
	"sync"
)

// 定义日志等级
const (
	DEBUG           = iota // 用于调度,最低级
	INFO                   //输出普通信息,常用
	WARNING                //输出警告,非错误信息,又比较重要
	ERROR                  //错误,属严重信息
	DEFAULT_FLAG    = log.LstdFlags
	LSHORTFILE_FLAG = log.Lshortfile | log.LstdFlags
)

// 定义日志接口
type ColorLogger interface {
	Debug(format string, s ...interface{})
	Info(format string, s ...interface{})
	Warn(format string, s ...interface{})
	Error(format string, s ...interface{})
	SetLevel(level int) //设置等级
}

//var isColorChan = make(chan bool, 1)

type Logger struct {
	Level   int //日志等级
	debug   *log.Logger
	info    *log.Logger
	warning *log.Logger
	error   *log.Logger
	IsColor bool   //是否使用带颜色的日志
	Depth   int    //详情深度
	Prefix  string //前缀
	wg      *sync.WaitGroup
}

// 对外初使日志
func New(w io.Writer, isColor bool) ColorLogger {
	return InitWriteLogger(w, 2, LSHORTFILE_FLAG, isColor)
}

// 默认的日志
var defaultLogger = InitWriteLogger(os.Stdout, 2, DEFAULT_FLAG, true)

// 初使写入日志,写入到一个buffer里
func InitWriteLogger(w io.Writer, depth int, flag int, isColor bool) ColorLogger {
	logger := new(Logger)
	logger.wg = new(sync.WaitGroup)
	logger.IsColor = isColor
	logger.Depth = depth
	//初使每个等级的日志
	logger.debug = log.New(w, logger.setColorString(DEBUG, "[DEBUG]"), flag)
	logger.info = log.New(w, logger.setColorString(INFO, "[INF]"), flag)
	logger.warning = log.New(w, logger.setColorString(WARNING, "[WAR]"), flag)
	logger.error = log.New(w, logger.setColorString(ERROR, "[ERR]"), flag)

	logger.SetLevel(DEBUG) //初使一下等级啦
	return logger
}

// 设置不同字体颜色
func (l *Logger) setColor(level int, format string, args ...interface{}) string {
	if false == l.IsColor {
		return fmt.Sprintf(format, args...)
	}
	switch level {
	case DEBUG:
		return gocolor.SMagenta(format, args...)
	case INFO:
		return gocolor.SGreen(format, args...)
	case WARNING:
		return gocolor.SYellow(format, args...)
	case ERROR:
		return gocolor.SRed(format, args...)
	default:
		return fmt.Sprintf(format, args...)
	}
}

// 设置不同背景颜色
func (l *Logger) setColorString(level int, format string, args ...interface{}) string {
	if false == l.IsColor {
		return fmt.Sprintf(format, args...)
	}
	switch level {
	case DEBUG:
		return gocolor.SMagentaBG(format, args...)
	case INFO:
		return gocolor.SGreenBG(format, args...)
	case WARNING:
		return gocolor.SYellowBG(format, args...)
	case ERROR:
		return gocolor.SRedBG(format, args...)
	default:
		return fmt.Sprintf(format, args...)
	}
}

// 设置等级,默认全输出
func (l *Logger) SetLevel(level int) {
	l.Level = level
}

// 用于调度的日志
func (l *Logger) Debug(format string, s ...interface{}) {
	if l.Level > DEBUG {
		return
	}
	l.debug.Output(l.Depth, l.setColor(DEBUG, format, s...))
}

// 输出普通信息
func (l *Logger) Info(format string, s ...interface{}) {
	if l.Level > INFO {
		return
	}
	l.info.Output(l.Depth, l.setColor(INFO, format, s...))
}

// 输出警告信息
func (l *Logger) Warn(format string, s ...interface{}) {
	if l.Level > WARNING {
		return
	}
	l.warning.Output(l.Depth, l.setColor(WARNING, format, s...))
}

// 输出错误信息
func (l *Logger) Error(format string, s ...interface{}) {
	if l.Level > ERROR {
		return
	}
	l.error.Output(l.Depth, l.setColor(ERROR, format, s...))
}

func Debug(format string, s ...interface{}) {
	defaultLogger.Debug(format, s...)
}
func Info(format string, s ...interface{}) {
	defaultLogger.Info(format, s...)
}
func Warn(format string, s ...interface{}) {
	defaultLogger.Warn(format, s...)
}
func Error(format string, s ...interface{}) {
	defaultLogger.Error(format, s...)
}
