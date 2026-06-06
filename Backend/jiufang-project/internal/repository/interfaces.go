package repository

import (
	"context"
	"time"

	"jiufang/internal/model/audit"
	"jiufang/internal/model/dialog"
	"jiufang/internal/model/export"
	"jiufang/internal/model/feedback"
	"jiufang/internal/model/permission"
	"jiufang/internal/model/query"
	"jiufang/internal/model/report"
	"jiufang/internal/model/user"
)

type UserRepositoryInterface interface {
	Create(ctx context.Context, u *user.User) error
	GetByID(ctx context.Context, id uint) (*user.User, error)
	GetBySnowflakeID(ctx context.Context, snowflakeID int64) (*user.User, error)
	GetByUsername(ctx context.Context, username string) (*user.User, error)
	GetByEmail(ctx context.Context, email string) (*user.User, error)
	List(ctx context.Context, offset, limit int, username, role string, status int) ([]user.User, int64, error)
	Update(ctx context.Context, u *user.User) error
	Delete(ctx context.Context, id uint) error
	UpdateStatus(ctx context.Context, id uint, status int) error
	UpdatePassword(ctx context.Context, id uint, password string) error
	UpdateAvatar(ctx context.Context, id uint, avatar string) error
	UpdateFirstLoginStatus(ctx context.Context, id uint, isFirstLogin bool) error
}

type UserGroupRepositoryInterface interface {
	Create(ctx context.Context, group *permission.UserGroup) error
	GetByID(ctx context.Context, id uint) (*permission.UserGroup, error)
	GetBySnowflakeID(ctx context.Context, snowflakeID int64) (*permission.UserGroup, error)
	GetByName(ctx context.Context, name string) (*permission.UserGroup, error)
	List(ctx context.Context, offset, limit int, name string) ([]permission.UserGroup, int64, error)
	Update(ctx context.Context, group *permission.UserGroup) error
	Delete(ctx context.Context, id uint) error
	GetMemberCount(ctx context.Context, groupID uint) (int64, error)
	AddMembers(ctx context.Context, groupID uint, userIDs []uint) error
	RemoveMembers(ctx context.Context, groupID uint, userIDs []uint) error
	GetMembers(ctx context.Context, groupID uint) ([]user.UserGroupMember, error)
	GetMembersWithPagination(ctx context.Context, groupID uint, offset, limit int) ([]user.UserGroupMember, int64, error)
	GetGroupsByUserID(ctx context.Context, userID uint) ([]permission.UserGroup, error)
	IsPresetGroup(ctx context.Context, snowflakeID int64) (bool, error)
}

type PermissionRepositoryInterface interface {
	Create(ctx context.Context, p *permission.Permission) error
	CreateBatch(ctx context.Context, permissions []permission.Permission) error
	GetByID(ctx context.Context, id uint) (*permission.Permission, error)
	GetBySnowflakeID(ctx context.Context, snowflakeID int64) (*permission.Permission, error)
	GetByGroupID(ctx context.Context, groupID uint) ([]permission.Permission, error)
	Update(ctx context.Context, p *permission.Permission) error
	Delete(ctx context.Context, id uint) error
	DeleteByGroupID(ctx context.Context, groupID uint) error
	List(ctx context.Context, offset, limit int, groupID uint, resourceType string) ([]permission.Permission, int64, error)
}

type DialogRepositoryInterface interface {
	Create(ctx context.Context, session *dialog.DialogSession) error
	GetByID(ctx context.Context, id uint) (*dialog.DialogSession, error)
	GetBySnowflakeID(ctx context.Context, snowflakeID string) (*dialog.DialogSession, error)
	GetByUserID(ctx context.Context, userID uint, offset, limit int) ([]dialog.DialogSession, int64, error)
	GetActiveSessionsByUserID(ctx context.Context, userID uint) ([]dialog.DialogSession, error)
	Update(ctx context.Context, session *dialog.DialogSession) error
	UpdateStatus(ctx context.Context, id uint, status string) error
	Delete(ctx context.Context, id uint) error
	CloseSession(ctx context.Context, snowflakeID string) error
}

// QueryRepositoryInterface defines the interface for query history and favorite operations.
type QueryRepositoryInterface interface {
	// QuerySession operations
	CreateQuerySession(ctx context.Context, session *query.QuerySession) error
	GetQuerySessionByID(ctx context.Context, id uint) (*query.QuerySession, error)
	GetQuerySessionBySnowflakeID(ctx context.Context, snowflakeID int64) (*query.QuerySession, error)
	GetQuerySessionsByUserID(ctx context.Context, userID int64, offset, limit int) ([]query.QuerySession, int64, error)
	CloseQuerySession(ctx context.Context, snowflakeID int64) error

	// QueryRecord operations
	CreateQueryRecord(ctx context.Context, record *query.QueryRecord) error
	GetQueryRecordByID(ctx context.Context, id uint) (*query.QueryRecord, error)
	GetQueryRecordBySnowflakeID(ctx context.Context, snowflakeID int64) (*query.QueryRecord, error)
	GetQueryRecordsByUserID(ctx context.Context, userID int64, offset, limit int, startTime, endTime string, status string) ([]query.QueryRecord, int64, error)
	GetQueryRecordsBySessionID(ctx context.Context, sessionID int64) ([]query.QueryRecord, error)
	DeleteQueryRecord(ctx context.Context, snowflakeID int64, userID int64) error

	// Favorite operations
	CreateFavorite(ctx context.Context, favorite *query.Favorite) error
	GetFavoriteByID(ctx context.Context, id uint) (*query.Favorite, error)
	GetFavoriteBySnowflakeID(ctx context.Context, snowflakeID int64) (*query.Favorite, error)
	GetFavoritesByUserID(ctx context.Context, userID int64, offset, limit int, name string) ([]query.Favorite, int64, error)
	DeleteFavorite(ctx context.Context, snowflakeID int64, userID int64) error
	IsFavoriteNameExists(ctx context.Context, userID int64, name string) (bool, error)
}

// DialogFavoriteRepositoryInterface defines the interface for dialog favorite operations.
type DialogFavoriteRepositoryInterface interface {
	// DialogFavorite operations
	CreateDialogFavorite(ctx context.Context, favorite *dialog.DialogFavorite) error
	GetDialogFavoriteByID(ctx context.Context, id uint) (*dialog.DialogFavorite, error)
	GetDialogFavoriteBySnowflakeID(ctx context.Context, snowflakeID int64) (*dialog.DialogFavorite, error)
	GetDialogFavoritesByUserID(ctx context.Context, userID int64, offset, limit int) ([]dialog.DialogFavorite, int64, error)
	DeleteDialogFavorite(ctx context.Context, snowflakeID int64, userID int64) error
	IsDialogFavoriteExists(ctx context.Context, userID int64, dialogSessionID int64) (bool, error)
	GetDialogFavoriteByUserAndSession(ctx context.Context, userID int64, dialogSessionID int64) (*dialog.DialogFavorite, error)
}

// ExportRecordRepositoryInterface defines the interface for export record operations.
type ExportRecordRepositoryInterface interface {
	// ExportRecord operations
	CreateExportRecord(ctx context.Context, record *export.ExportRecord) error
	GetExportRecordByID(ctx context.Context, id uint) (*export.ExportRecord, error)
	GetExportRecordBySnowflakeID(ctx context.Context, snowflakeID int64) (*export.ExportRecord, error)
	GetExportRecordsByUserID(ctx context.Context, userID int64, offset, limit int) ([]export.ExportRecord, int64, error)
}

// ReportRepositoryInterface defines the interface for scheduled report and push record operations.
type ReportRepositoryInterface interface {
	// ScheduledReport operations
	Create(ctx context.Context, scheduledReport *report.ScheduledReport) error
	GetByID(ctx context.Context, id uint) (*report.ScheduledReport, error)
	GetBySnowflakeID(ctx context.Context, snowflakeID int64) (*report.ScheduledReport, error)
	List(ctx context.Context, offset, limit int, name, status string) ([]report.ScheduledReport, int64, error)
	Update(ctx context.Context, snowflakeID int64, updates map[string]interface{}) error
	Delete(ctx context.Context, snowflakeID int64) error
	GetActiveReports(ctx context.Context) ([]report.ScheduledReport, error)

	// PushRecord operations
	CreatePushRecord(ctx context.Context, pushRecord *report.PushRecord) error
	GetPushRecordByID(ctx context.Context, id uint) (*report.PushRecord, error)
	GetPushRecordBySnowflakeID(ctx context.Context, snowflakeID int64) (*report.PushRecord, error)
	GetPushRecordsByReportID(ctx context.Context, reportID int64, offset, limit int, pushStatus string) ([]report.PushRecord, int64, error)
	ListPushRecords(ctx context.Context, offset, limit int, pushType string, pushStatus string, startTime, endTime string) ([]report.PushRecord, int64, error)
	UpdatePushRecord(ctx context.Context, snowflakeID int64, updates map[string]interface{}) error
}

// FeedbackRepositoryInterface defines the interface for feedback operations.
type FeedbackRepositoryInterface interface {
	// Feedback operations
	Create(ctx context.Context, feedback *feedback.Feedback) error
	GetByID(ctx context.Context, id uint) (*feedback.Feedback, error)
	GetBySnowflakeID(ctx context.Context, snowflakeID int64) (*feedback.Feedback, error)
	GetByQueryRecordID(ctx context.Context, queryRecordID int64) (*feedback.Feedback, error)
	GetByUserID(ctx context.Context, userID int64, offset, limit int) ([]feedback.Feedback, int64, error)
	List(ctx context.Context, offset, limit int, userID int64, rating string) ([]feedback.Feedback, int64, error)
	IsFeedbackExists(ctx context.Context, queryRecordID int64) (bool, error)
}

// AlertRepositoryInterface defines the interface for alert rule operations.
type AlertRepositoryInterface interface {
	// Alert operations
	Create(ctx context.Context, alert *report.Alert) error
	GetByID(ctx context.Context, id uint) (*report.Alert, error)
	GetBySnowflakeID(ctx context.Context, snowflakeID int64) (*report.Alert, error)
	List(ctx context.Context, offset, limit int, name string, status string) ([]report.Alert, int64, error)
	Update(ctx context.Context, snowflakeID int64, updates map[string]interface{}) error
	Delete(ctx context.Context, snowflakeID int64) error
	GetActiveAlerts(ctx context.Context) ([]report.Alert, error)
	UpdateLastTriggeredAt(ctx context.Context, snowflakeID int64, triggeredAt time.Time) error
}

// OperationLogRepositoryInterface defines the interface for operation log operations.
type OperationLogRepositoryInterface interface {
	// OperationLog operations
	Create(ctx context.Context, log *audit.OperationLog) error
	GetByID(ctx context.Context, id uint) (*audit.OperationLog, error)
	GetBySnowflakeID(ctx context.Context, snowflakeID int64) (*audit.OperationLog, error)
	List(ctx context.Context, offset, limit int, userID int64, operationType string, startTime, endTime string) ([]audit.OperationLog, int64, error)
	GetByUsername(ctx context.Context, username string) (int64, error)
}
