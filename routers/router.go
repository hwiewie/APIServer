package routers

import (
	"github.com/gin-gonic/gin"
	/* 	"github.com/jamsa/gin-k8s/api"
	   	admv1 "github.com/jamsa/gin-k8s/api/v1/admin"
	   	k8sv1 "github.com/jamsa/gin-k8s/api/v1/k8s"
	   	_ "github.com/jamsa/gin-k8s/docs"
	   	"github.com/jamsa/gin-k8s/pkg/logging"
	   	"github.com/jamsa/gin-k8s/pkg/setting" */
	// "github.com/swaggo/gin-swagger"
)

// InitRouter initialize routing information
func InitRouter() *gin.Engine {
	r := gin.New()
	/* 	if setting.AppSetting.LogGin {
	   		r.Use(logging.LogToLogrus())
	   	}

	   	r.Use(gin.Recovery())

	   	r.POST("/auth", api.GetAuth)
	   	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	   	apiv1 := r.Group("/api/v1")
	   	//apiv1.Use(jwt.JWT())
	   	{
	   		apiv1.GET("/k8s/:cluster/pods", k8sv1.GetPods)
	   		apiv1.GET("/k8s/:cluster/pods/:namespace/:podname", k8sv1.GetPod)

	   		apiv1.GET("/k8s/:cluster/deployments", k8sv1.GetDeployments)
	   		apiv1.GET("/k8s/:cluster/deployments/:namespace/:deploymentName", k8sv1.GetDeployment)

	   		apiv1.GET("/k8s/:cluster/services", k8sv1.GetServices)
	   		apiv1.GET("/k8s/:cluster/services/:namespace/:serviceName", k8sv1.GetService)

	   		apiv1.GET("/k8s/:cluster/statefulsets", k8sv1.GetStatefulSets)
	   		apiv1.GET("/k8s/:cluster/statefulsets/:namespace/:statefulsetName", k8sv1.GetStatefulSet)

	   		apiv1.GET("/k8s/:cluster/ingresses", k8sv1.GetIngresses)
	   		apiv1.GET("/k8s/:cluster/ingresses/:namespace/:ingressName", k8sv1.GetIngress)

	   		apiv1.GET("/k8s/:cluster/configmaps", k8sv1.GetConfigMaps)
	   		apiv1.GET("/k8s/:cluster/configmaps/:namespace/:configmapName", k8sv1.GetConfigMap)

	   		apiv1.GET("/k8s/:cluster/persistentvolumeclaims", k8sv1.GetPersistentVolumeClaims)
	   		apiv1.GET("/k8s/:cluster/persistentvolumeclaims/:namespace/:persistentvolumeclaimName", k8sv1.GetPersistentVolumeClaim)

	   		//不区分namespace的
	   		apiv1.GET("/k8s/:cluster/persistentvolumes", k8sv1.GetPersistentVolumes)
	   		apiv1.GET("/k8s/:cluster/persistentvolumes/:persistentvolumeName", k8sv1.GetPersistentVolume)

	   		apiv1.GET("/k8s/:cluster/nodes", k8sv1.GetNodes)
	   		apiv1.GET("/k8s/:cluster/nodes/:nodeName", k8sv1.GetNode)

	   		apiv1.GET("/k8s/:cluster/namespaces", k8sv1.GetNamespaces)
	   		apiv1.GET("/k8s/:cluster/namespaces/:namespaceName", k8sv1.GetNamespace)

	   		//集群
	   		apiv1.POST("/admin/clusters/query", admv1.GetClusters)
	   		apiv1.GET("/admin/clusters/:id", admv1.GetCluster)
	   		apiv1.POST("/admin/clusters", admv1.AddCluster)
	   		apiv1.PUT("/admin/clusters/:id", admv1.EditCluster)
	   		apiv1.DELETE("/admin/clusters/:id", admv1.DeleteCluster)

	   		//用户
	   		apiv1.POST("/admin/users/query", admv1.GetUsers)
	   		apiv1.GET("/admin/users/:id", admv1.GetUser)
	   		apiv1.POST("/admin/users", admv1.AddUser)
	   		apiv1.PUT("/admin/users/:id", admv1.EditUser)
	   		apiv1.DELETE("/admin/users/:id", admv1.DeleteUser)
	   	} */
	return r
}
