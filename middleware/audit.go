package middleware

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type auditEntry struct {
	Time     string `json:"time"`
	ClientIP string `json:"client_ip"`
	UserID   string `json:"user_id"`
	Role     string `json:"role"`
	Method   string `json:"method"`
	Path     string `json:"path"`
	Status   int    `json:"status"`
}

type AuditLogger struct {
	mu   sync.Mutex
	file *os.File
}

func NewAuditLogger(path string) (*AuditLogger, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	return &AuditLogger{file: file}, nil
}

func (l *AuditLogger) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		user := CurrentUser(c)
		entry := auditEntry{
			Time:     time.Now().UTC().Format(time.RFC3339),
			ClientIP: c.ClientIP(),
			UserID:   user.ID,
			Role:     string(user.Role),
			Method:   c.Request.Method,
			Path:     c.Request.URL.Path,
			Status:   c.Writer.Status(),
		}
		if err := l.write(entry); err != nil {
			log.Printf("error de auditoría: %v", err)
		}
	}
}

func (l *AuditLogger) write(entry auditEntry) error {
	line, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	_, err = l.file.Write(append(line, '\n'))
	return err
}