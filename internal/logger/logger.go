package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"

	"wishlist/internal/config"
	"wishlist/internal/utils/colors"
)

var (
	logFile      *os.File
	currentLevel = LevelInfo
	consoleLog   *log.Logger
	fileLog      *log.Logger
	jsonFileLog  *log.Logger
)

type Level int

const (
	LevelError Level = iota
	LevelWarn
	LevelInfo
	LevelDebug
)

type jsonEntry struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

func SetupLogger() {
	fmt.Print("Setting up logger...")

	switch viper.GetString(config.LogLevel) {
	case "DEBUG":
		currentLevel = LevelDebug
	case "ERROR":
		currentLevel = LevelError
	case "WARN":
		currentLevel = LevelWarn
	default:
		currentLevel = LevelInfo
	}

	if viper.GetBool(config.LogToFile) {
		filePath := viper.GetString(config.LogFilePath)

		flags := os.O_CREATE | os.O_WRONLY | os.O_APPEND
		if viper.GetString(config.LogFileMode) == "overwrite" {
			flags = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
		}
		if viper.GetString(config.LogFileMode) == "rotate" {
			rotateLogFile(filePath, viper.GetString(config.LogFilesFolder))
		}

		var err error
		logFile, err = os.OpenFile(filePath, flags, 0666)
		if err != nil {
			fmt.Println()
			log.Fatalf("Fatal: failed to open log file: %v", err)
		}

		if viper.GetString(config.LogFormat) == "json" {
			jsonFileLog = log.New(logFile, "", 0)
			b, _ := json.Marshal(jsonEntry{Time: ts(), Level: "INFO", Message: "=== new run ==="})
			jsonFileLog.Println(string(b))
		} else {
			fileLog = log.New(logFile, "", 0)
			_, _ = logFile.WriteString("\n\n==== New run at " + time.Now().Format("2006/01/02 15:04:05") + " ====\n")
		}
	}

	if viper.GetBool(config.LogToConsole) {
		consoleLog = log.New(os.Stdout, "", 0)
	}

	fmt.Println(colors.Green("      Done."))
}

func rotateLogFile(filePath, logsFolder string) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return
	}

	if err := os.MkdirAll(logsFolder, 0755); err != nil {
		log.Printf("failed to create logs folder: %v", err)
		return
	}

	base := filepath.Base(filePath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	newName := filepath.Join(logsFolder, name+"_"+time.Now().Format("2006-01-02_15-04-05")+ext)
	if err := os.Rename(filePath, newName); err != nil {
		log.Printf("failed to rotate log file: %v", err)
	}
}

func GetWriters() io.Writer {
	if jsonFileLog != nil {
		if consoleLog != nil {
			return os.Stdout
		}
		return io.Discard
	}
	switch {
	case consoleLog != nil && fileLog != nil:
		return io.MultiWriter(os.Stdout, logFile)
	case fileLog != nil:
		return logFile
	case consoleLog != nil:
		return os.Stdout
	default:
		return io.Discard
	}
}

func GetLevel() Level { return currentLevel }

func SetLevel(l Level) { currentLevel = l }

func Debug(format string, args ...any) {
	if currentLevel < LevelDebug {
		return
	}
	write(colors.Gray("DEBUG"), "DEBUG", fmt.Sprintf(format, args...))
}

func Info(format string, args ...any) {
	if currentLevel < LevelInfo {
		return
	}
	write(colors.Green("INFO "), "INFO", fmt.Sprintf(format, args...))
}

func Warn(format string, args ...any) {
	if currentLevel < LevelWarn {
		return
	}
	write(colors.Yellow("WARN "), "WARN", colors.Yellow(fmt.Sprintf(format, args...)))
}

func Error(format string, args ...any) {
	write(colors.Red("ERROR"), "ERROR", colors.Red(fmt.Sprintf(format, args...)))
}

func ErrorWithID(ctx context.Context, format string, args ...any) {
	reqID, _ := ctx.Value("request_id").(string)
	Error("[%s] "+colors.Red(format), append([]any{reqID}, args...)...)
}

func Fatal(v ...any) {
	write(colors.Bold(colors.Red("FATAL")), "FATAL", fmt.Sprint(v...))
	os.Exit(1)
}

func Fatalf(format string, args ...any) {
	write(colors.Bold(colors.Red("FATAL")), "FATAL", fmt.Sprintf(format, args...))
	os.Exit(1)
}

func write(coloredLevel, plainLevel, text string) {
	if consoleLog != nil {
		consoleLog.Printf("%s %s %s", ts(), coloredLevel, text)
	}
	if fileLog != nil {
		fileLog.Printf("%s %s %s", ts(), plainLevel, text)
	}
	if jsonFileLog != nil {
		b, _ := json.Marshal(jsonEntry{Time: ts(), Level: plainLevel, Message: text})
		jsonFileLog.Println(string(b))
	}
}

func ts() string { return time.Now().Format("2006/01/02 15:04:05") }

type GlobalLogger struct{} // Wraps package-level functions for use as an injected Logger dependency in service

func (GlobalLogger) Error(format string, v ...any) { Error(format, v...) }
