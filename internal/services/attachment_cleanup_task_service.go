package services

import (
	"sync"

	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/sqls"

	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
)

type AttachmentCleanupTaskStatus struct {
	Running    bool   `json:"running"`
	Processed  int64  `json:"processed"`
	Total      int64  `json:"total"`
	StartedAt  int64  `json:"startedAt"`
	FinishedAt int64  `json:"finishedAt"`
	Error      string `json:"error"`
}

type attachmentCleanupTaskService struct {
	mu     sync.Mutex
	status AttachmentCleanupTaskStatus
}

var AttachmentCleanupTaskService = &attachmentCleanupTaskService{}

func (s *attachmentCleanupTaskService) Status() AttachmentCleanupTaskStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.status
}

func (s *attachmentCleanupTaskService) Start(before int64) (AttachmentCleanupTaskStatus, bool) {
	s.mu.Lock()
	if s.status.Running {
		status := s.status
		s.mu.Unlock()
		return status, false
	}
	var total int64
	sqls.DB().Model(&models.Attachment{}).Where("topic_id = ? AND status = ? AND create_time < ?", 0, constants.StatusOk, before).Count(&total)
	s.status = AttachmentCleanupTaskStatus{Running: true, Total: total, StartedAt: dates.NowTimestamp()}
	status := s.status
	s.mu.Unlock()
	go s.run(before)
	return status, true
}

func (s *attachmentCleanupTaskService) run(before int64) {
	processed, err := AttachmentService.CleanupOrphans(before)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.Running = false
	s.status.Processed = processed
	s.status.FinishedAt = dates.NowTimestamp()
	if err != nil {
		s.status.Error = err.Error()
	}
}
