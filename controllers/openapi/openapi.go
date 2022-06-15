// Package openapi Wayne OpenAPI Document
//
// wayne 開放 API （以下簡稱 openapi）是一組便於開發者調試、第三方工具開發和 CI/CD 的開放數據接口。
// openapi 雖然格式上滿足 Restful，但是並不是單一接口只針對特定資源的操作，在大部分時候單一接口會操作一組資源；
// 同時，雖然 openapi 下只允許通過 GET 請求訪問，但是並不意味著 GET 操作代表著 Restful 中對 GET 的用法定義；
// openapi 的路徑格式：/openapi/v1/gateway/action/:action，:action 代表特定操作，例如： get_vip_info、upgrade_deployment。
//
// openapi 所操作的 action 必須搭配具有該 action 權限的 APIKey 使用（作為一個命名為 apikey 的 GET 請求參數），
// 而對應的 apikey 需要具備 action 對應的權限（例如：action 對應 get_pod_info 的時候，apikey 需要具備 OPENAPI_GET_POD_INFO 權限），
// 同時，受限於某些action的使用場景，可能強制要求附加的 APIKey 的使用範圍，目前APIKey的適用範圍包括三種，App 級別、Namespace 級別 以及 全局級別。
// Terms Of Service:
//
//
//     Schemes: https
//     Host: localhost
//     BasePath: /openapi/v1/gateway/action
//     Version: 1.6.1
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
//     Security:
//     - api_key:
//
//     SecurityDefinitions:
//     api_key:
//          type: apiKey
//          name: apikey
//          in: query
//
// swagger:meta
package openapi

import (
	"fmt"
	"net/http"

	"github.com/hwiewie/APIServer/controllers/base"
	"github.com/hwiewie/APIServer/models"
	"github.com/hwiewie/APIServer/util/logs"
)

const (
	GetPodInfoAction             = "GET_POD_INFO"
	GetPodInfoFromIPAction       = "GET_POD_INFO_FROM_IP"
	GetResourceInfoAction        = "GET_RESOURCE_INFO"
	GetDeploymentStatusAction    = "GET_DEPLOYMENT_STATUS"
	UpgradeDeploymentAction      = "UPGRADE_DEPLOYMENT"
	ScaleDeploymentAction        = "SCALE_DEPLOYMENT"
	RestartDeploymentAction      = "RESTART_DEPLOYMENT"
	GetDeploymentDetailAction    = "GET_DEPLOYMENT_DETAIL"
	GetLatestDeploymentTplAction = "GET_LATEST_DEPLOYMENT_TPL"
	GetPodListAction             = "GET_POD_LIST"

	ListNamespaceUsers = "LIST_NAMESPACE_USERS"
	ListNamespaceApps  = "LIST_NAMESPACE_APPS"
	ListAppDeploys     = "List_APP_DEPLOYS"

	PermissionPrefix = "OPENAPI_"
)

type OpenAPIController struct {
	base.APIKeyController
}

func (c *OpenAPIController) Prepare() {
	c.APIKeyController.Prepare()
}

func (c *OpenAPIController) CheckoutRoutePermission(action string) bool {
	permission := false
	for _, p := range c.APIKey.Group.Permissions {
		if p.Name == PermissionPrefix+action {
			permission = true
		}
	}
	if !permission {
		c.AddErrorAndResponse(fmt.Sprintf("APIKey does not have the following permission: %s", PermissionPrefix+action), http.StatusUnauthorized)
		return false
	}
	return true
}

func (c *OpenAPIController) CheckDeploymentPermission(deployment string) bool {
	if c.APIKey.Type == models.NamespaceAPIKey {
		d, err := models.DeploymentModel.GetByName(deployment)
		if err != nil {
			logs.Error("Failed to get deployment by name", err)
			c.AddErrorAndResponse("", http.StatusBadRequest)
			return false
		}
		app, _ := models.AppModel.GetById(d.AppId)
		if app.Namespace.Id != c.APIKey.ResourceId {
			c.AddErrorAndResponse(fmt.Sprintf("APIKey does not have permission to operate request resource: %s", deployment), http.StatusUnauthorized)
			return false
		}
	}
	if c.APIKey.Type == models.ApplicationAPIKey {
		deploy, err := models.DeploymentModel.GetByName(deployment)
		if err != nil {
			logs.Error("Failed to get deployment by name", err)
			c.AddErrorAndResponse("", http.StatusBadRequest)
			return false
		}
		if deploy.AppId != c.APIKey.ResourceId {
			c.AddErrorAndResponse(fmt.Sprintf("APIKey does not have permission to operate request resource(deployment): %s", deployment), http.StatusUnauthorized)
			return false
		}
	}
	return true
}

func (c *OpenAPIController) CheckNamespacePermission(namespace string) bool {
	if c.APIKey.Type == models.NamespaceAPIKey {
		ns, err := models.NamespaceModel.GetByName(namespace)
		if err != nil {
			logs.Error("Failed to get namespace by name", err)
			c.AddErrorAndResponse("", http.StatusBadRequest)
			return false
		}
		if ns.Deleted == true {
			c.AddErrorAndResponse(fmt.Sprintf("The requested namespace has been offline: %s", namespace), http.StatusBadRequest)
			return false
		}
		if ns.Id != c.APIKey.ResourceId {
			c.AddErrorAndResponse(fmt.Sprintf("APIKey does not have permission to operate request resource(namespace): %s", namespace), http.StatusUnauthorized)
			return false
		}
	}
	return true
}