package models

// AnnouncementPublishRecord keeps the audit trail for site notice changes.
type AnnouncementPublishRecord struct {
	Model
	OperatorId      int64  `gorm:"not null;index:idx_announcement_record_operator" json:"operatorId" form:"operatorId"`
	PreviousContent string `gorm:"type:text" json:"previousContent" form:"previousContent"`
	Content         string `gorm:"type:text;not null" json:"content" form:"content"`
	Status          string `gorm:"size:32;not null;index:idx_announcement_record_status" json:"status" form:"status"`
	Error           string `gorm:"type:text" json:"error" form:"error"`
	PublishTime     int64  `gorm:"not null;index:idx_announcement_record_publish_time" json:"publishTime" form:"publishTime"`
	CreateTime      int64  `gorm:"not null" json:"createTime" form:"createTime"`
}

// DashboardViewPreference stores named dashboard views per administrator.
type DashboardViewPreference struct {
	Model
	UserId     int64  `gorm:"not null;uniqueIndex:uk_dashboard_view_preference;index:idx_dashboard_view_preference_user" json:"userId" form:"userId"`
	ViewKey    string `gorm:"size:256;not null;uniqueIndex:uk_dashboard_view_preference" json:"viewKey" form:"viewKey"`
	Views      string `gorm:"type:text;not null" json:"views" form:"views"`
	UpdateTime int64  `gorm:"not null" json:"updateTime" form:"updateTime"`
}
