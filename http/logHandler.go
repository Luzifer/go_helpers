package http

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Luzifer/go_helpers/v2/accessLogger"
	"github.com/sirupsen/logrus"
)

type HTTPLogHandler struct {
	Handler          http.Handler
	TrustedIPHeaders []string

	logger logrus.StdLogger
}

func NewHTTPLogHandler(h http.Handler) http.Handler {
	return NewHTTPLogHandlerWithLogger(h, log.New(os.Stderr, "", log.LstdFlags))
}

func NewHTTPLogHandlerWithLogger(h http.Handler, l logrus.StdLogger) http.Handler {
	return HTTPLogHandler{
		Handler:          h,
		TrustedIPHeaders: []string{"X-Forwarded-For", "RemoteAddr", "X-Real-IP"},
		logger:           l,
	}
}

func (l HTTPLogHandler) ServeHTTP(res http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ares := accessLogger.New(res)

	l.Handler.ServeHTTP(ares, r)

	path := r.URL.Path
	if q := r.URL.Query().Encode(); len(q) > 0 {
		path = path + "?" + q
	}

	l.logger.Printf("%s - \"%s %s\" %d %d \"%s\" \"%s\" %s",
		l.findIP(r),
		r.Method,
		path,
		ares.StatusCode,
		ares.Size,
		r.Header.Get("Referer"),
		r.Header.Get("User-Agent"),
		time.Since(start),
	)
}

func (l HTTPLogHandler) findIP(r *http.Request) string {
	remoteAddr := strings.SplitN(r.RemoteAddr, ":", 2)[0]

	for _, hdr := range l.TrustedIPHeaders {
		if value := r.Header.Get(hdr); value != "" {
			return strings.SplitN(value, ",", 2)[0]
		}
	}

	return remoteAddr
}
