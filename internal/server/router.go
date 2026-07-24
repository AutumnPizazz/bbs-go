package server

import (
	adminHandlers "bbs-go/internal/handlers/admin"
	apiHandlers "bbs-go/internal/handlers/api"
	"bbs-go/internal/middleware"
	"bbs-go/internal/pkg/config"
	"bbs-go/internal/pkg/ginx"
	"bbs-go/internal/pkg/respath"
	"bbs-go/internal/services"
	webspa "bbs-go/web"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/web"
	"github.com/spf13/cast"
)

func NewServer() {
	printBanner()
	if err := newRouter().Run(":" + cast.ToString(config.Instance.Port)); err != nil {
		slog.Error(err.Error(), slog.Any("err", err))
		os.Exit(-1)
	}
}

func newRouter() *gin.Engine {
	conf := config.Instance
	if conf == nil {
		conf = &config.Config{}
	}

	gin.SetMode(gin.ReleaseMode)
	app := gin.New()
	app.Use(gin.Recovery())
	app.Use(gin.Logger())
	corsConfig := cors.Config{
		AllowOrigins:     conf.AllowedOrigins,
		AllowCredentials: true,
		MaxAge:           600,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodOptions, http.MethodHead, http.MethodDelete, http.MethodPut},
		AllowHeaders:     []string{"*"},
	}
	if len(corsConfig.AllowOrigins) == 0 {
		corsConfig.AllowAllOrigins = true
		corsConfig.AllowCredentials = false
	}
	app.Use(cors.New(corsConfig))
	app.Use(middleware.AttachmentMiddleware)

	registerAPIRoutes(app.Group("/api", middleware.InstallMiddleware, middleware.AuthMiddleware))
	registerAdminRoutes(app.Group("/api/admin", middleware.InstallMiddleware, middleware.AuthMiddleware, middleware.AdminMiddleware, middleware.CSRFMiddleware))

	app.StaticFS("/res", ginx.StaticFiles(respath.ResDir()))
	ginx.HandleSPA(app, ginx.SPAOptions{
		Root:         "./web/build/spa",
		EmbeddedFS:   webspa.SPA,
		EmbeddedRoot: "build/spa",
		DirOptions: ginx.DirOptions{
			ShowList:  false,
			SPA:       true,
			IndexName: "index.html",
		},
		NotFoundPrefixes: []string{"/api/", "/res/"},
		NotFoundHandler: func(ctx *gin.Context) {
			ginx.WriteHttpStatusJSON(ctx, http.StatusNotFound, web.JsonErrorCode(http.StatusNotFound, "Not found"))
		},
	})
	app.GET("/sitemap.xml", middleware.AuthMiddleware, middleware.ContentAccessMiddleware, func(ctx *gin.Context) {
		redirectURL := services.SeoSitemapService.RedirectURL()
		if strs.IsBlank(redirectURL) {
			ginx.WriteHttpStatusJSON(ctx, http.StatusNotFound, web.JsonErrorCode(http.StatusNotFound, "Not found"))
			return
		}
		ctx.Redirect(http.StatusFound, redirectURL)
	})

	return app
}

func registerAPIRoutes(group *gin.RouterGroup) {
	installGroup := group.Group("/install")
	installGroup.GET("/status", apiHandlers.InstallStatus)
	installGroup.POST("/test_db_connection", apiHandlers.InstallTestDbConnection)
	installGroup.POST("/install", apiHandlers.InstallInstall)

	topicGroup := group.Group("/topic", middleware.ContentAccessMiddleware)
	topicGroup.GET("/category_navs", apiHandlers.CategoryNavs)
	topicGroup.GET("/categories", apiHandlers.Categories)
	topicGroup.GET("/category", apiHandlers.Category)
	topicGroup.POST("/create", apiHandlers.TopicCreate)
	topicGroup.GET("/edit/:id", apiHandlers.TopicEditForm)
	topicGroup.POST("/edit/:id", apiHandlers.TopicEdit)
	topicGroup.POST("/delete/:id", apiHandlers.TopicRemove)
	topicGroup.POST("/recommend/:id", apiHandlers.TopicRecommend)
	topicGroup.GET("/recentlikes/:id", apiHandlers.TopicRecentlikes)
	topicGroup.GET("/recent", apiHandlers.TopicRecent)
	topicGroup.GET("/user_topics", apiHandlers.TopicUserTopics)
	topicGroup.GET("/topics", apiHandlers.TopicTopics)
	topicGroup.POST("/accept_answer/:id", apiHandlers.TopicAcceptAnswer)
	topicGroup.POST("/unaccept_answer/:id", apiHandlers.TopicUnacceptAnswer)
	topicGroup.GET("/tag/topics", apiHandlers.TopicTagTopics)
	topicGroup.GET("/favorite/:id", apiHandlers.TopicFavorite)
	topicGroup.POST("/sticky/:id", apiHandlers.TopicSticky)
	topicGroup.GET("/hide_content", apiHandlers.TopicHideContent)
	topicGroup.GET("/:id", apiHandlers.TopicDetail)

	loginGroup := group.Group("/login")
	loginGroup.POST("/signin", apiHandlers.LoginSignin)
	loginGroup.GET("/signout", apiHandlers.LoginSignout)
	loginGroup.GET("/wx_login_config", apiHandlers.LoginWxLoginConfig)
	loginGroup.POST("/wx_login_submit", apiHandlers.LoginWxLoginSubmit)
	loginGroup.POST("/wx_bind", apiHandlers.LoginWxBind)
	loginGroup.POST("/wx_unbind", apiHandlers.LoginWxUnbind)
	loginGroup.GET("/google_login_config", apiHandlers.LoginGoogleLoginConfig)
	loginGroup.POST("/google_login_submit", apiHandlers.LoginGoogleLoginSubmit)
	loginGroup.POST("/google_bind", apiHandlers.LoginGoogleBind)
	loginGroup.POST("/google_one_tap", apiHandlers.LoginGoogleOneTap)
	loginGroup.POST("/google_unbind", apiHandlers.LoginGoogleUnbind)
	loginGroup.GET("/github_login_config", apiHandlers.LoginGithubLoginConfig)
	loginGroup.POST("/github_login_submit", apiHandlers.LoginGithubLoginSubmit)
	loginGroup.POST("/github_unbind", apiHandlers.LoginGithubUnbind)

	userGroup := group.Group("/user")
	userGroup.GET("/current", apiHandlers.UserCurrent)
	userGroup.POST("/update/:id", apiHandlers.UserUpdate)
	userGroup.POST("/update_avatar", apiHandlers.UserUpdateAvatar)
	userGroup.POST("/set_username", apiHandlers.UserSetUsername)
	userGroup.POST("/set_background_image", apiHandlers.UserSetBackgroundImage)
	userGroup.GET("/favorites", apiHandlers.UserFavorites)
	userGroup.GET("/msg_recent", apiHandlers.UserMsgRecent)
	userGroup.GET("/messages", apiHandlers.UserMessages)
	userGroup.POST("/forbidden", apiHandlers.UserForbidden)
	userGroup.GET("/wx_bind_info", apiHandlers.UserWxBindInfo)
	userGroup.GET("/google_bind_info", apiHandlers.UserGoogleBindInfo)
	userGroup.GET("/github_bind_info", apiHandlers.UserGithubBindInfo)
	userGroup.GET("/:id", apiHandlers.UserDetail)

	tagGroup := group.Group("/tag", middleware.ContentAccessMiddleware)
	tagGroup.GET("/tags", apiHandlers.TagTags)
	tagGroup.POST("/autocomplete", apiHandlers.TagAutocompleteSubmit)
	tagGroup.GET("/:id", apiHandlers.TagDetail)

	commentGroup := group.Group("/comment", middleware.ContentAccessMiddleware)
	commentGroup.GET("/comments", apiHandlers.CommentComments)
	commentGroup.GET("/replies", apiHandlers.CommentReplies)
	commentGroup.POST("/create", apiHandlers.CommentCreate)
	commentGroup.POST("/delete/:id", apiHandlers.CommentRemove)

	favoriteGroup := group.Group("/favorite", middleware.ContentAccessMiddleware)
	favoriteGroup.POST("/add", apiHandlers.FavoriteAdd)
	favoriteGroup.POST("/delete", apiHandlers.FavoriteRemove)

	likeGroup := group.Group("/like", middleware.ContentAccessMiddleware)
	likeGroup.POST("/like", apiHandlers.LikeLike)
	likeGroup.POST("/unlike", apiHandlers.LikeUnlike)
	likeGroup.GET("/liked_ids", apiHandlers.LikeLikedIds)
	likeGroup.GET("/liked", apiHandlers.LikeLiked)

	configGroup := group.Group("/config")
	configGroup.GET("/configs", apiHandlers.ConfigConfigs)
	configGroup.GET("/about", apiHandlers.ConfigAbout)

	uploadGroup := group.Group("/upload")
	uploadGroup.POST("", apiHandlers.UploadHandle)

	attachmentGroup := group.Group("/attachment", middleware.ContentAccessMiddleware)
	attachmentGroup.POST("/upload", apiHandlers.AttachmentUpload)
	attachmentGroup.GET("/download/:id", apiHandlers.AttachmentDownload)

	linkGroup := group.Group("/link")
	linkGroup.GET("/list", apiHandlers.LinkList)
	linkGroup.GET("/top_links", apiHandlers.LinkTopLinks)

	captchaGroup := group.Group("/captcha")
	captchaGroup.GET("/request", apiHandlers.CaptchaRequest)
	captchaGroup.GET("/verify", apiHandlers.CaptchaVerify)
	captchaGroup.GET("/request_angle", apiHandlers.CaptchaRequestAngle)

	searchGroup := group.Group("/search", middleware.ContentAccessMiddleware)
	searchGroup.GET("/topic", apiHandlers.SearchTopic)
	searchGroup.GET("/user", apiHandlers.SearchUser)

	fansGroup := group.Group("/fans")
	fansGroup.POST("/follow", apiHandlers.FansFollow)
	fansGroup.POST("/unfollow", apiHandlers.FansUnfollow)
	fansGroup.GET("/is_followed", apiHandlers.FansIsFollowed)
	fansGroup.GET("/fans", apiHandlers.FansFans)
	fansGroup.GET("/followed", apiHandlers.FansFollowed)
	fansGroup.GET("/recent/fans", apiHandlers.FansRecentFans)
	fansGroup.GET("/recent/follow", apiHandlers.FansRecentFollow)

	userReportGroup := group.Group("/user-report")
	userReportGroup.POST("/submit", apiHandlers.UserReportSubmit)

	voteGroup := group.Group("/vote", middleware.ContentAccessMiddleware)
	voteGroup.POST("/cast", apiHandlers.VoteCast)
	voteGroup.GET("/:id", apiHandlers.VoteDetail)

}

func registerAdminRoutes(group *gin.RouterGroup) {
	roleGroup := group.Group("/role")
	roleGroup.GET("/roles", adminHandlers.Roles)
	roleGroup.POST("/list", adminHandlers.RoleList)
	roleGroup.GET("/permissions", adminHandlers.RolePermissions)
	roleGroup.POST("/create", adminHandlers.RoleCreate)
	roleGroup.POST("/update", adminHandlers.RoleUpdate)
	roleGroup.POST("/update_permissions", adminHandlers.RoleUpdatePermissions)
	roleGroup.POST("/delete", adminHandlers.RoleRemove)
	roleGroup.POST("/update_sort", adminHandlers.RoleUpdateSort)
	roleGroup.GET("/:id", adminHandlers.RoleDetail)

	dictTypeGroup := group.Group("/dict-type")
	dictTypeGroup.GET("/list", adminHandlers.DictTypeList)
	dictTypeGroup.POST("/list", adminHandlers.DictTypeList)
	dictTypeGroup.POST("/create", adminHandlers.DictTypeCreate)
	dictTypeGroup.POST("/update", adminHandlers.DictTypeUpdate)
	dictTypeGroup.POST("/delete", adminHandlers.DictTypeRemove)
	dictTypeGroup.GET("/:id", adminHandlers.DictTypeDetail)

	dictGroup := group.Group("/dict")
	dictGroup.GET("/list", adminHandlers.DictList)
	dictGroup.POST("/create", adminHandlers.DictCreate)
	dictGroup.POST("/update", adminHandlers.DictUpdate)
	dictGroup.POST("/delete", adminHandlers.DictRemove)
	dictGroup.POST("/update_sort", adminHandlers.DictUpdateSort)
	dictGroup.GET("/dicts", adminHandlers.DictDicts)
	dictGroup.GET("/:id", adminHandlers.DictDetail)

	commonGroup := group.Group("/common")
	commonGroup.GET("/overview", adminHandlers.CommonOverview)
	commonGroup.GET("/preferences/views", adminHandlers.DashboardViewPreferences)
	commonGroup.POST("/preferences/views", adminHandlers.SaveDashboardViewPreferences)

	healthGroup := group.Group("/health")
	healthGroup.GET("", adminHandlers.Health)

	taskGroup := group.Group("/tasks")
	taskGroup.GET("/status", adminHandlers.TaskStatus)

	backupGroup := group.Group("/backup")
	backupGroup.GET("/config", adminHandlers.DatabaseBackupConfig)
	backupGroup.POST("/config", adminHandlers.DatabaseBackupSaveConfig)
	backupGroup.GET("/list", adminHandlers.DatabaseBackupList)
	backupGroup.POST("/create", adminHandlers.DatabaseBackupCreate)
	backupGroup.POST("/restore", adminHandlers.DatabaseBackupRestore)
	backupGroup.POST("/delete", adminHandlers.DatabaseBackupRemove)
	backupGroup.GET("/download/:id", adminHandlers.DatabaseBackupDownload)

	commentGroup := group.Group("/comment")
	commentGroup.POST("/list", adminHandlers.CommentList)
	commentGroup.POST("/delete", adminHandlers.CommentRemove)
	commentGroup.GET("/:id", adminHandlers.CommentDetail)
	commentGroup.DELETE("/:id", adminHandlers.CommentRemove)

	userGroup := group.Group("/user")
	userGroup.GET("/synccount", adminHandlers.UserSynccount)
	userGroup.POST("/list", adminHandlers.UserList)
	userGroup.POST("/create", adminHandlers.UserCreate)
	userGroup.POST("/update", adminHandlers.UserUpdate)
	userGroup.POST("/forbidden", adminHandlers.UserForbidden)
	userGroup.POST("/batch/preview", adminHandlers.UserBatchPreview)
	userGroup.POST("/batch", adminHandlers.UserBatch)
	userGroup.POST("/update_password", adminHandlers.UserUpdatePassword)
	userGroup.POST("/reset_password", adminHandlers.UserResetPassword)
	userGroup.GET("/:id", adminHandlers.UserDetail)

	tagGroup := group.Group("/tag")
	tagGroup.POST("/list", adminHandlers.TagList)
	tagGroup.POST("/create", adminHandlers.TagCreate)
	tagGroup.POST("/update", adminHandlers.TagUpdate)
	tagGroup.GET("/autocomplete", adminHandlers.TagAutocomplete)
	tagGroup.GET("/tags", adminHandlers.TagTags)
	tagGroup.GET("/:id", adminHandlers.TagDetail)

	attachmentGroup := group.Group("/attachment")
	attachmentGroup.POST("/list", adminHandlers.AttachmentList)
	attachmentGroup.POST("/delete", adminHandlers.AttachmentRemove)
	attachmentGroup.POST("/cleanup-orphans", adminHandlers.AttachmentCleanupOrphans)
	attachmentGroup.GET("/:id", adminHandlers.AttachmentDetail)

	messageGroup := group.Group("/message")
	messageGroup.POST("/list", adminHandlers.MessageList)
	messageGroup.POST("/create", adminHandlers.MessageCreate)
	messageGroup.POST("/update", adminHandlers.MessageUpdate)
	messageGroup.POST("/delete", adminHandlers.MessageRemove)
	messageGroup.POST("/task/preview", adminHandlers.MessageTaskPreview)
	messageGroup.POST("/task/create", adminHandlers.MessageTaskCreate)
	messageGroup.POST("/task/list", adminHandlers.MessageTaskList)
	messageGroup.POST("/task/send", adminHandlers.MessageTaskSend)
	messageGroup.POST("/task/retry", adminHandlers.MessageTaskRetry)
	messageGroup.GET("/task/:id", adminHandlers.MessageTaskDetail)
	messageGroup.GET("/:id", adminHandlers.MessageDetail)

	announcementGroup := group.Group("/announcement")
	announcementGroup.GET("/history", adminHandlers.AnnouncementHistory)
	announcementGroup.POST("/preview", adminHandlers.AnnouncementPreview)
	announcementGroup.POST("/publish", adminHandlers.AnnouncementPublish)

	favoriteGroup := group.Group("/favorite")
	favoriteGroup.POST("/list", adminHandlers.FavoriteList)
	favoriteGroup.POST("/create", adminHandlers.FavoriteCreate)
	favoriteGroup.POST("/update", adminHandlers.FavoriteUpdate)
	favoriteGroup.GET("/:id", adminHandlers.FavoriteDetail)

	topicGroup := group.Group("/topic")
	topicGroup.POST("/list", adminHandlers.TopicList)
	topicGroup.POST("/recommend", adminHandlers.TopicRecommend)
	topicGroup.DELETE("/recommend", adminHandlers.TopicRemoveRecommend)
	topicGroup.POST("/sticky", adminHandlers.TopicSticky)
	topicGroup.POST("/batch/preview", adminHandlers.TopicBatchPreview)
	topicGroup.POST("/batch", adminHandlers.TopicBatch)
	topicGroup.POST("/delete", adminHandlers.TopicRemove)
	topicGroup.POST("/undelete", adminHandlers.TopicUndelete)
	topicGroup.POST("/accept_answer", adminHandlers.TopicAcceptAnswer)
	topicGroup.POST("/unaccept_answer", adminHandlers.TopicUnacceptAnswer)
	topicGroup.POST("/mark_solved", adminHandlers.TopicMarkSolved)
	topicGroup.POST("/mark_unsolved", adminHandlers.TopicMarkUnsolved)
	topicGroup.GET("/:id", adminHandlers.TopicDetail)

	categoryGroup := group.Group("/category")
	categoryGroup.POST("/list", adminHandlers.CategoryList)
	categoryGroup.POST("/create", adminHandlers.CategoryCreate)
	categoryGroup.POST("/update", adminHandlers.CategoryUpdate)
	categoryGroup.GET("/options", adminHandlers.CategoryOptions)
	categoryGroup.POST("/update_sort", adminHandlers.CategoryUpdateSort)
	categoryGroup.POST("/delete", adminHandlers.CategoryRemove)
	categoryGroup.GET("/:id", adminHandlers.CategoryDetail)

	sysConfigGroup := group.Group("/sys-config")
	sysConfigGroup.POST("/list", adminHandlers.SysConfigList)
	sysConfigGroup.GET("/configs", adminHandlers.SysConfigConfigs)
	sysConfigGroup.POST("/save", adminHandlers.SysConfigSave)
	sysConfigGroup.POST("/save-sensitive", adminHandlers.SysConfigSaveSensitive)
	sysConfigGroup.GET("/:id", adminHandlers.SysConfigDetail)

	searchGroup := group.Group("/search")
	searchGroup.GET("/reindex/status", adminHandlers.SearchReindexStatus)
	searchGroup.POST("/reindex", adminHandlers.SearchReindex)

	seoGroup := group.Group("/seo")
	seoGroup.GET("/sitemap/status", adminHandlers.SeoSitemapStatus)
	seoGroup.POST("/sitemap/generate", adminHandlers.SeoSitemapGenerate)

	linkGroup := group.Group("/link")
	linkGroup.POST("/list", adminHandlers.LinkList)
	linkGroup.POST("/create", adminHandlers.LinkCreate)
	linkGroup.POST("/update", adminHandlers.LinkUpdate)
	linkGroup.POST("/delete", adminHandlers.LinkRemove)
	linkGroup.POST("/update_sort", adminHandlers.LinkUpdateSort)
	linkGroup.GET("/:id", adminHandlers.LinkDetail)

	operateLogGroup := group.Group("/operate-log")
	operateLogGroup.POST("/list", adminHandlers.OperateLogList)
	operateLogGroup.GET("/:id", adminHandlers.OperateLogDetail)

	userReportGroup := group.Group("/user-report")
	userReportGroup.POST("/list", adminHandlers.UserReportList)
	userReportGroup.POST("/create", adminHandlers.UserReportCreate)
	userReportGroup.POST("/update", adminHandlers.UserReportUpdate)
	userReportGroup.POST("/process", adminHandlers.UserReportProcess)
	userReportGroup.POST("/action", adminHandlers.UserReportAction)
	userReportGroup.POST("/batch/preview", adminHandlers.UserReportBatchPreview)
	userReportGroup.POST("/batch", adminHandlers.UserReportBatch)
	userReportGroup.GET("/:id", adminHandlers.UserReportDetail)

	voteGroup := group.Group("/vote")
	voteGroup.POST("/list", adminHandlers.VoteList)
	voteGroup.POST("/create", adminHandlers.VoteCreate)
	voteGroup.POST("/update", adminHandlers.VoteUpdate)
	voteGroup.POST("/delete", adminHandlers.VoteRemove)
	voteGroup.GET("/:id", adminHandlers.VoteDetail)

	voteOptionGroup := group.Group("/vote-option")
	voteOptionGroup.POST("/list", adminHandlers.VoteOptionList)
	voteOptionGroup.POST("/create", adminHandlers.VoteOptionCreate)
	voteOptionGroup.POST("/update", adminHandlers.VoteOptionUpdate)
	voteOptionGroup.POST("/delete", adminHandlers.VoteOptionRemove)
	voteOptionGroup.GET("/:id", adminHandlers.VoteOptionDetail)

	voteRecordGroup := group.Group("/vote-record")
	voteRecordGroup.POST("/list", adminHandlers.VoteRecordList)
	voteRecordGroup.POST("/create", adminHandlers.VoteRecordCreate)
	voteRecordGroup.POST("/update", adminHandlers.VoteRecordUpdate)
	voteRecordGroup.POST("/delete", adminHandlers.VoteRecordRemove)
	voteRecordGroup.GET("/:id", adminHandlers.VoteRecordDetail)

}
