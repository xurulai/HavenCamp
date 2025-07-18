package ssl

import (
	"haven_camp_server/pkg/zlog"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/unrolled/secure"
)

func TlsHandler(host string, port int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查SSL证书文件是否存在
		certFile := "/etc/ssl/certs/server.crt"
		keyFile := "/etc/ssl/private/server.key"

		sslRedirect := false
		if _, err := os.Stat(certFile); err == nil {
			if _, err := os.Stat(keyFile); err == nil {
				sslRedirect = true
			}
		}

		secureMiddleware := secure.New(secure.Options{
			SSLRedirect: sslRedirect,
			SSLHost:     host + ":" + strconv.Itoa(port),
		})
		err := secureMiddleware.Process(c.Writer, c.Request)

		// If there was an error, do not continue.
		if err != nil {
			zlog.Fatal(err.Error())
			return
		}

		c.Next()
	}
}
