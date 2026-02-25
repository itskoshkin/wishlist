package ginutils

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	"wishlist/internal/utils/colors"
	"wishlist/internal/utils/ua"
)

const (
	pathWidth    = 50
	pathMaxWidth = 80
)

func CustomGinLogger(out io.Writer) gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			ipAddr, referer := ExtractIPAndReferer(param)
			requestID, _ := param.Keys["request_id"].(string)

			path := param.Path
			if utf8.RuneCountInString(path) > pathMaxWidth {
				path = string([]rune(path)[:pathMaxWidth-3]) + "..."
			}

			if utf8.RuneCountInString(path) > pathWidth {
				return twoLinedAccessLog(param, path)
			}

			return fmt.Sprintf("%s %s [%s] | %7s %-50s | %s | %12v | %-15s | %s%s\n",
				param.TimeStamp.Format("2006/01/02 15:04:05"),
				colors.Purple("GIN  "),
				colors.Gray(requestID),
				param.Method,
				path,
				mapHTTPCodeToColor(param.StatusCode, strconv.Itoa(param.StatusCode)),
				param.Latency,
				ipAddr,
				useragent.FormatUserAgent(param.Request.UserAgent()),
				referer,
			)
		},
		Output: out,
	})
}

func twoLinedAccessLog(param gin.LogFormatterParams, path string) string {
	ipAddr, referer := ExtractIPAndReferer(param)
	requestID, _ := param.Keys["request_id"].(string)

	prefix := fmt.Sprintf("%s %s [%s] | %7s ",
		param.TimeStamp.Format("2006/01/02 15:04:05"),
		colors.Purple("GIN  "),
		colors.Gray(requestID),
		param.Method)
	visiblePrefixLen := 39 + utf8.RuneCountInString(requestID)

	firstLine := prefix + path + "\n"
	secondLine := fmt.Sprintf("%s| %s | %12v | %-15s | %s%s\n",
		strings.Repeat(" ", visiblePrefixLen+pathWidth+1),
		mapHTTPCodeToColor(param.StatusCode, strconv.Itoa(param.StatusCode)),
		param.Latency,
		ipAddr,
		useragent.FormatUserAgent(param.Request.UserAgent()),
		referer,
	)
	return firstLine + secondLine
}

func mapHTTPCodeToColor(code int, message string) string {
	switch {
	case code >= 200 && code < 300:
		return colors.Green(message)
	case code >= 400 && code < 500:
		return colors.Yellow(message)
	case code >= 500:
		return colors.Red(message)
	default:
		return message
	}
}
