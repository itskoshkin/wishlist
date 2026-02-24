package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	"wishlist/internal/utils/gin"
	"wishlist/internal/utils/ua"
)

var (
	logFile     *os.File
	logFileOnce sync.Once
)

func GetLogFile() *os.File {
	logFileOnce.Do(func() {
		var err error
		logFile, err = os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Fatal: %v", err)
		}
	})
	return logFile
}

func SetupLogger() {
	file := GetLogFile()
	log.SetOutput(io.MultiWriter(os.Stdout, file))
	log.SetFlags(log.Ldate | log.Ltime)
	_, _ = file.WriteString("==== New run at " + time.Now().Format("2006-01-02 15:04:05") + " ====\n")
}

func CustomGinLogger(out io.Writer) gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			ipAddr, referer := ginutils.ExtractIPAndReferer(param)
			requestID, _ := param.Keys["request_id"].(string)

			if utf8.RuneCountInString(param.Path) > ginutils.PathWidth {
				return ginutils.TwoLinedAccessLog(param)
			}

			return fmt.Sprintf("[GIN] [%s] %s | %7s %-42s | %3d | %10v | %-15s | %s%s\n",
				requestID,
				param.TimeStamp.Format("2006/01/02 - 15:04:05"),
				param.Method,
				param.Path,
				param.StatusCode,
				param.Latency,
				ipAddr,
				useragent.FormatUserAgent(param.Request.UserAgent()),
				referer,
			)
		},
		Output: out,
	})
}
