package auth

import (
	"log/slog"
	"time"
)

// StartLoginLogCleanup 启动登录日志定期清理，每 24 小时删除 90 天前的日志。
func (s *Service) StartLoginLogCleanup() {
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			s.cleanupLoginLogs()
		}
	}()
}

func (s *Service) cleanupLoginLogs() {
	before := time.Now().AddDate(0, 0, -90)
	deleted, err := s.repo.DeleteLoginLogsBefore(before)
	if err != nil {
		slog.Warn("login log cleanup failed", "error", err)
		return
	}
	if deleted > 0 {
		slog.Info("login log cleanup", "deleted", deleted, "before", before.Format(time.DateOnly))
	}
}
