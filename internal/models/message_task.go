package models

const (
	MessageTaskDraft     = 0
	MessageTaskRunning   = 1
	MessageTaskCompleted = 2
	MessageTaskFailed    = 3

	MessageDeliveryPending = 0
	MessageDeliverySent    = 1
	MessageDeliveryFailed  = 2
)

// MessageSendTask is an auditable broadcast job independent from inbox rows.
type MessageSendTask struct {
	Model
	CreatorId   int64  `gorm:"not null;index:idx_message_task_creator" json:"creatorId" form:"creatorId"`
	FromId      int64  `gorm:"not null" json:"fromId" form:"fromId"`
	Title       string `gorm:"size:1024;not null" json:"title" form:"title"`
	Content     string `gorm:"type:text;not null" json:"content" form:"content"`
	TargetType  string `gorm:"size:32;not null;index:idx_message_task_target" json:"targetType" form:"targetType"`
	TargetId    int64  `gorm:"not null;default:0;index:idx_message_task_target" json:"targetId" form:"targetId"`
	Status      int    `gorm:"not null;default:0;index:idx_message_task_status" json:"status" form:"status"`
	TotalCount  int64  `gorm:"not null;default:0" json:"totalCount" form:"totalCount"`
	SentCount   int64  `gorm:"not null;default:0" json:"sentCount" form:"sentCount"`
	FailedCount int64  `gorm:"not null;default:0" json:"failedCount" form:"failedCount"`
	Error       string `gorm:"type:text" json:"error" form:"error"`
	StartedAt   int64  `json:"startedAt" form:"startedAt"`
	FinishedAt  int64  `json:"finishedAt" form:"finishedAt"`
	CreateTime  int64  `gorm:"not null;index:idx_message_task_create_time" json:"createTime" form:"createTime"`
	UpdateTime  int64  `gorm:"not null" json:"updateTime" form:"updateTime"`
}

// MessageDelivery records the outcome for each recipient in a broadcast job.
type MessageDelivery struct {
	Model
	TaskId     int64  `gorm:"not null;index:idx_message_delivery_task" json:"taskId" form:"taskId"`
	UserId     int64  `gorm:"not null;index:idx_message_delivery_user" json:"userId" form:"userId"`
	MessageId  int64  `gorm:"not null;default:0" json:"messageId" form:"messageId"`
	Status     int    `gorm:"not null;default:0;index:idx_message_delivery_status" json:"status" form:"status"`
	Attempts   int    `gorm:"not null;default:0" json:"attempts" form:"attempts"`
	Error      string `gorm:"type:text" json:"error" form:"error"`
	CreateTime int64  `gorm:"not null" json:"createTime" form:"createTime"`
	UpdateTime int64  `gorm:"not null" json:"updateTime" form:"updateTime"`
}
