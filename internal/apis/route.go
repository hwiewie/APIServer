package apis

import (
	"github.com/gin-gonic/gin"
	//"github.com/hwiewie/APIServer/pkg/confgin"
)

// RootGroup 向 httpserver 註冊根路由
func RootGroup(base *gin.RouterGroup) {
	v0 := base.Group("v0")
	v0.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	// 向 v0 註冊 Deployments 路由
	// confgin.AppendGroup(v0, DeploymentRouterGroup)
	// confgin.AppendGroup(v0, PodRouterGroup)
	// confgin.AppendGroup(v0, ServiceRouterGroup)
	// confgin.AppendGroup(v0, IngressRouterGroup)
}