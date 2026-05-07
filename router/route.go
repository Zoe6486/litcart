package router

import (
	"litcart/logger"
	"litcart/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// func Setup(userCtrl *controller.UserController, commCtrl *controller.CommunityController) *gin.Engine {
func Setup(db *sqlx.DB) *gin.Engine {
	r := gin.New()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})
	// 1. 第一位：生成 Request ID (给后面所有人用)
	r.Use(middleware.RequestIDMiddleware())
	// 2. 第二位：记录日志 ,异常捕获，分开写更好吗？
	r.Use(logger.GinLogger(), logger.GinRecovery(true))
	// //加载前端
	// r.LoadHTMLFiles("./templates/index.html")
	// r.Static("/static", "./static")
	// r.GET("/", func(c *gin.Context) {
	// 	c.HTML(http.StatusOK, "index.html", nil)
	// })
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "litcart api server",
		})
	})

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// ========================
	// Dependency Injection
	// ========================
	// 依赖注入 (Wire Up)
	// 这里是核心：手动组装对象树
	// ── Wire up DAOs ──────────────────────────────────────────────────────
	userDao := mysql.NewUserDao(db)
	// communityDao := mysql.NewCommunityDao(db)
	// postDao := mysql.NewPostDao(db)
	// postVoteDao := mysql.NewPostVoteDao(db)
	// postCommentDao := mysql.NewPostCommentDao(db)

	// ── Wire up Logic ─────────────────────────────────────────────────────
	userLogic := logic.NewUserLogic(userDao)
	// communityLogic := logic.NewCommunityLogic(communityDao)
	// postLogic := logic.NewPostLogic(postDao)
	// voteLogic := logic.NewPostVoteLogic(postVoteDao, postDao)
	// commentLogic := logic.NewPostCommentLogic(postCommentDao, postDao)

	// ── Wire up Controllers ───────────────────────────────────────────────
	userCtrl := controller.NewUserController(userLogic)
	// communityCtrl := controller.NewCommunityController(communityLogic)
	// postCtrl := controller.NewPostController(postLogic)
	// voteCtrl := controller.NewPostVoteController(voteLogic)
	// commentCtrl := controller.NewPostCommentController(commentLogic)

	// v1 := r.Group("/api/v1")
	api := r.Group("/api")
	v1 := api.Group("/v1")

	// // 注册
	// v1.POST("/signup", controller.SignUpHandler)
	// // 登录
	// v1.POST("/login", controller.LoginHandler)
	v1.POST("/signup", userCtrl.SignUpHandler)
	v1.POST("/login", userCtrl.LoginHandler)

	// // community routes
	// community := v1.Group("/communities")
	// {
	// 	community.GET("", communityCtrl.GetCommunityListHandler)
	// 	community.GET("/:id", communityCtrl.GetCommunityDetailHandler)
	// }

	// // Posts  (auth middleware goes here when ready)
	// posts := v1.Group("/posts")
	// // posts.Use(middleware.JWTAuth())
	// posts.Use(middleware.AuthMiddleware())
	// {
	// 	posts.POST("", postCtrl.CreatePost)
	// 	posts.GET("", postCtrl.ListPosts)
	// 	posts.GET("/:post_id", postCtrl.GetPost)
	// 	posts.DELETE("/:post_id", postCtrl.DeletePost)

	// 	// Vote — POST /api/v1/posts/vote
	// 	posts.POST("/vote", voteCtrl.VotePost)

	// 	// Comments — nested under a specific post
	// 	comments := posts.Group("/:post_id/comments")
	// 	{
	// 		comments.GET("", commentCtrl.ListComments)
	// 		comments.POST("", commentCtrl.CreateComment)
	// 		comments.DELETE("/:comment_id", commentCtrl.DeleteComment)
	// 	}
	// }

	return r
}
