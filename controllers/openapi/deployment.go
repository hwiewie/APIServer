package openapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"

	"github.com/hwiewie/APIServer/client"
	"github.com/hwiewie/APIServer/controllers/common"
	"github.com/hwiewie/APIServer/models"
	"github.com/hwiewie/APIServer/models/response"
	resdeployment "github.com/hwiewie/APIServer/resources/deployment"
	"github.com/hwiewie/APIServer/resources/pod"
	"github.com/hwiewie/APIServer/util/hack"
	"github.com/hwiewie/APIServer/util/logs"
)

type DeploymentInfo struct {
	Deployment         *models.Deployment
	DeploymentTemplete *models.DeploymentTemplate
	DeploymentObject   *appsv1.Deployment
	Cluster            *models.Cluster
	Namespace          *models.Namespace
}

// swagger:parameters DeploymentStatusParam
type DeploymentStatusParam struct {
	// in: query
	// Required: true
	Deployment string `json:"deployment"`
	// Required: true
	Namespace string `json:"namespace"`
	// 和升級部署存在差別，不允許同時填寫多個 cluster
	// Required: true
	Cluster string `json:"cluster"`
}

// swagger:parameters RestartDeploymentParam
type RestartDeploymentParam struct {
	// in: query
	// Required: true
	Deployment string `json:"deployment"`
	// Required: true
	Namespace string `json:"namespace"`
	// 和升級部署存在差別，不允許同時填寫多個 cluster
	// Required: true
	Cluster string `json:"cluster"`
}

// swagger:parameters UpgradeDeploymentParam
type UpgradeDeploymentParam struct {
	// in: query
	// Required: true
	Deployment string `json:"deployment"`
	// Required: true
	Namespace string `json:"namespace"`
	// 支持同時填寫多個 Cluster，只需要在 cluster 之間使用英文半角的逗號分隔即可
	// Required: true
	Cluster  string `json:"cluster"`
	clusters []string
	// Required: false
	TemplateId int `json:"template_id"`
	// 該字段為 true 的時候，會自動使用新生成的配置模板上線，否則會只創建對應的模板，並且將模板 ID 返回（用於敏感的需要手動操作的上線環境）
	// Required: false
	Publish bool `json:"publish"`
	// 升級的描述
	// Required: false
	Description string `json:"description"`
	// 該字段為扁平化為字符串的 key-value 字典，填寫格式為 容器名1=鏡像名1,容器名2=鏡像名2 (即:多個容器之間使用英文半角的逗號分隔）
	// Required: false
	Images   string `json:"images"`
	imageMap map[string]string
	// 該字段為扁平化為字符串的 key-value 字典，填寫格式為 環境變量1=值1,環境變量2=值2 (即:多個環境變量之間使用英文半角的逗號分隔）
	// Required: false
	Environments string `json:"environments"`
	envMap       map[string]string
}

// swagger:parameters ScaleDeploymentParam
type ScaleDeploymentParam struct {
	// in: query
	// Required: true
	Deployment string `json:"deployment"`
	// Required: true
	Namespace string `json:"namespace"`
	// 和升級部署存在差別，不允許同時填寫多個 cluster
	// Required: true
	Cluster string `json:"cluster"`
	// 期望調度到的副本數量，範圍：(0,32]
	// Required: true
	Replicas int `json:"replicas"`
}

// swagger:model deploymentstatus
type DeploymentStatus struct {
	// required: true
	Pods []response.Pod `json:"pods"`
	// required: true
	Deployment response.Deployment `json:"deployment"`
	// required: true
	Healthz bool `json:"healthz"`
}

// 重點關注 kubernetes 集群內狀態而非描述信息，當然也可以只關注 healthz 字段
// swagger:response respdeploymentstatus
type respdeploymentstatus struct {
	// in: body
	// Required: true
	Body struct {
		response.ResponseBase
		DeploymentStatus DeploymentStatus `json:"status"`
	}
}

// swagger:parameters DeploymentStatusParam
type deploymentDetailParam struct {
	// in: query
	// Required: true
	Namespace string `json:"namespace"`
	// Required: true
	App string `json:"app"`
	// Required: true
	Deployment string `json:"deployment"`
}

// swagger:response deploymentDetail
type deploymentDetailResponse struct {
	// in: body
	// Required: true
	Body struct {
		response.ResponseBase
		Deployment *models.Deployment `json:"deployment"`
	}
}

// swagger:parameters LatestDeploymentTplParam
type latestDeploymentTplParam struct {
	// in: query
	// Required: true
	Namespace string `json:"namespace"`
	// Required: true
	App string `json:"app"`
	// Required: true
	Deployment string `json:"deployment"`
}

// swagger:response deploymentDetail
type latestDeploymentTplResponse struct {
	// in: body
	// Required: true
	Body struct {
		response.ResponseBase
		DeploymentTpl *models.DeploymentTemplate `json:"deployment_tpl"`
	}
}

// swagger:route GET /get_deployment_status deploy DeploymentStatusParam
//
// 該接口用於返回特定部署的狀態信息
//
// 重點關注 kubernetes 集群內狀態而非描述信息，當然也可以只關注 healthz 字段。
// 該接口可以使用所有種類的 apikey
//
//     Responses:
//       200: respdeploymentstatus
//       400: responseState
//       401: responseState
//       500: responseState
// @router /get_deployment_status [get]
func (c *OpenAPIController) GetDeploymentStatus() {
	param := DeploymentStatusParam{
		c.GetString("deployment"),
		c.GetString("namespace"),
		c.GetString("cluster"),
	}
	if !c.CheckoutRoutePermission(GetDeploymentStatusAction) {
		return
	}
	if !c.CheckDeploymentPermission(param.Deployment) {
		return
	}
	if !c.CheckNamespacePermission(param.Namespace) {
		return
	}

	var result respdeploymentstatus // 返回數據的結構體
	result.Body.Code = http.StatusOK
	ns, err := models.NamespaceModel.GetByName(param.Namespace)
	if err != nil {
		logs.Error("Failed get namespace by name", param.Namespace, err)
		c.AddErrorAndResponse(fmt.Sprintf("Failed get namespace by name(%s)", param.Namespace), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal([]byte(ns.MetaData), &ns.MetaDataObj)
	if err != nil {
		logs.Error(fmt.Sprintf("Failed to parse metadata: %s", err.Error()))
		c.AddErrorAndResponse("", http.StatusInternalServerError)
		return
	}
	manager, err := client.Manager(param.Cluster)
	if err == nil {
		deployInfo, err := resdeployment.GetDeploymentDetail(manager.Client, manager.CacheFactory, param.Deployment, ns.KubeNamespace)
		if err != nil {
			logs.Error("Failed to get  k8s deployment state: %s", err.Error())
			c.AddErrorAndResponse("", http.StatusInternalServerError)
			return
		}
		result.Body.DeploymentStatus.Deployment = response.Deployment{
			Name:       deployInfo.ObjectMeta.Name,
			Namespace:  deployInfo.ObjectMeta.Namespace,
			Labels:     deployInfo.ObjectMeta.Labels,
			CreateTime: deployInfo.ObjectMeta.CreationTimestamp.Time,
			PodsState: response.PodInfo{
				Current:   deployInfo.Pods.Current,
				Desired:   deployInfo.Pods.Desired,
				Running:   deployInfo.Pods.Running,
				Pending:   deployInfo.Pods.Pending,
				Failed:    deployInfo.Pods.Failed,
				Succeeded: deployInfo.Pods.Succeeded,
			},
		}
		for _, e := range deployInfo.Pods.Warnings {
			result.Body.DeploymentStatus.Deployment.PodsState.Warnings = append(result.Body.DeploymentStatus.Deployment.PodsState.Warnings, fmt.Sprint(e))
		}
		podInfo, err := pod.GetPodsByDeployment(manager.CacheFactory, ns.KubeNamespace, param.Deployment)
		if err != nil {
			logs.Error("Failed to get k8s pod state: %s", err.Error())
			c.AddErrorAndResponse("", http.StatusInternalServerError)
			return
		}
		for _, pod := range podInfo {
			p := response.Pod{
				Name:      pod.Name,
				Namespace: pod.Namespace,
				State:     pod.State,
				PodIp:     pod.PodIp,
				NodeName:  pod.NodeName,
				StartTime: &pod.StartTime,
				Labels:    pod.Labels,
			}
			for _, status := range pod.ContainerStatus {
				p.ContainerStatus = append(p.ContainerStatus, response.ContainerStatus{status.Name, status.RestartCount})
			}
			result.Body.DeploymentStatus.Pods = append(result.Body.DeploymentStatus.Pods, p)
		}

		result.Body.DeploymentStatus.Healthz = true
		if deployInfo.Pods.Current != deployInfo.Pods.Desired {
			result.Body.DeploymentStatus.Healthz = false
		}

		for _, p := range podInfo {
			if p.PodIp == "" || p.State != string(v1.PodRunning) {
				result.Body.DeploymentStatus.Healthz = false
			}
		}
		c.HandleResponse(result.Body)
		return
	} else {
		logs.Error("Failed to get k8s client list", err)
		c.AddErrorAndResponse("Failed to get k8s client list!", http.StatusInternalServerError)
		return
	}
}

// swagger:route GET /restart_deployment deploy RestartDeploymentParam
//
// 用於用戶調用以實現強制重啟部署
//
// 該接口只能使用 app 級別的 apikey，這樣做的目的主要是防止 apikey 的濫用
//
//     Responses:
//       200: responseSuccess
//       400: responseState
//       401: responseState
//       500: responseState
// @router /restart_deployment [get]
func (c *OpenAPIController) RestartDeployment() {
	param := RestartDeploymentParam{
		Deployment: c.GetString("deployment"),
		Namespace:  c.GetString("namespace"),
		Cluster:    c.GetString("cluster"),
	}
	if !c.CheckoutRoutePermission(RestartDeploymentAction) || !c.CheckDeploymentPermission(param.Deployment) || !c.CheckNamespacePermission(param.Namespace) {
		return
	}
	if len(param.Namespace) == 0 {
		c.AddErrorAndResponse(fmt.Sprintf("Invalid namespace parameter"), http.StatusBadRequest)
		return
	}
	if len(param.Deployment) == 0 {
		c.AddErrorAndResponse(fmt.Sprintf("Invalid deployment parameter"), http.StatusBadRequest)
		return
	}
	ns, err := models.NamespaceModel.GetByName(param.Namespace)
	if err != nil {
		c.AddErrorAndResponse(fmt.Sprintf("Failed get namespace by name(%s)", param.Namespace), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal([]byte(ns.MetaData), &ns.MetaDataObj)
	if err != nil {
		logs.Error(fmt.Sprintf("Failed to parse metadata: %s", err.Error()))
		c.AddErrorAndResponse("", http.StatusInternalServerError)
		return
	}
	deployResource, err := models.DeploymentModel.GetByName(param.Deployment)
	if err != nil {
		c.AddErrorAndResponse(fmt.Sprintf("Failed get deployment by name(%s)", param.Deployment), http.StatusBadRequest)
		return
	}

	cli, err := client.Client(param.Cluster)
	if err != nil {
		logs.Error("Failed to connect to k8s client", err)
		c.AddErrorAndResponse(fmt.Sprintf("Failed to connect to k8s client on %s!", param.Cluster), http.StatusInternalServerError)
		return
	}

	deployObj, err := resdeployment.GetDeployment(cli, param.Deployment, ns.KubeNamespace)
	if err != nil {
		logs.Error("Failed to get deployment from k8s client", err.Error())
		c.AddErrorAndResponse(fmt.Sprintf("Failed to get deployment from k8s client on %s!", param.Cluster), http.StatusInternalServerError)
		return
	}
	deployObj.Spec.Template.ObjectMeta.Labels["timestamp"] = strconv.FormatInt(time.Now().Unix(), 10)

	if err := updateDeployment(deployObj, param.Cluster, c.APIKey.String(), "Restart Deployment", deployResource.Id); err != nil {
		logs.Error("Failed to restart from k8s client", err.Error())
		c.AddErrorAndResponse(fmt.Sprintf("Failed to restart from k8s client on %s!", param.Cluster), http.StatusInternalServerError)
		return
	}
	c.HandleResponse(nil)
}

// swagger:route GET /upgrade_deployment deploy UpgradeDeploymentParam
//
// 用於 CI/CD 中的集成升級部署
//
// 該接口只能使用 app 級別的 apikey，這樣做的目的主要是防止 apikey 的濫用。
// 目前用戶可以選擇兩種用法，第一種是默認的，會根據請求的 images 和 environments 對特定部署線上模板進行修改並創建新模板，然後使用新模板進行升級；
// 需要說明的是，environments 列表會對 deployment 內所有容器中包含指定環境變量 key 的環境變量進行更新，如不包含，則不更新。
// 第二種是通過指定 publish=false 來關掉直接上線，這種條件下會根據 images 和 environments 字段創建新的模板，並返回新模板id，用戶可以選擇去平台上手動上線或者通過本接口指定template_id參數上線。
// cluster 字段可以選擇單個機房也可以選擇多個機房，對於創建模板並上線的用法，會根據指定的機房之前的模板進行分類（如果機房 a 和機房 b 使用同一個模板，那麼調用以後仍然共用一個新模板）
// 而對於指定 template_id 來上線的形式，則會忽略掉所有檢查，直接使用特定模板上線到所有機房。
//
//     Responses:
//       200: responseSuccess
//       400: responseState
//       401: responseState
//       500: responseState
// @router /upgrade_deployment [get]
func (c *OpenAPIController) UpgradeDeployment() {
	param := UpgradeDeploymentParam{
		Deployment:   c.GetString("deployment"),
		Namespace:    c.GetString("namespace"),
		Cluster:      c.GetString("cluster"),
		Description:  c.GetString("description"),
		Images:       c.GetString("images"),
		Environments: c.GetString("environments"),
	}
	if !c.CheckoutRoutePermission(UpgradeDeploymentAction) || !c.CheckDeploymentPermission(param.Deployment) || !c.CheckNamespacePermission(param.Namespace) {
		return
	}
	param.clusters = strings.Split(param.Cluster, ",")
	var err error
	param.Publish, err = c.GetBool("publish", true)
	if err != nil {
		c.AddErrorAndResponse(fmt.Sprintf("Invalid publish parameter: %s", err.Error()), http.StatusBadRequest)
		return
	}
	param.TemplateId, err = c.GetInt("template_id", 0)
	if err != nil {
		c.AddErrorAndResponse(fmt.Sprintf("Invalid template_id parameter: %s", err.Error()), http.StatusBadRequest)
		return
	}

	// 根據特定模板升級，無須拼湊
	if param.TemplateId != 0 && param.Publish {
		for _, cluster := range param.clusters {
			deployInfo, err := getOnlineDeploymenetInfo(param.Deployment, param.Namespace, cluster, int64(param.TemplateId))
			if err != nil {
				logs.Error("Failed to get online deployment", err)
				c.AddError(fmt.Sprintf("Failed to get online deployment on %s!", cluster))
				continue
			}
			common.DeploymentPreDeploy(deployInfo.DeploymentObject, deployInfo.Deployment, deployInfo.Cluster, deployInfo.Namespace)
			err = publishDeployment(deployInfo, c.APIKey.String())
			if err != nil {
				logs.Error("Failed to publish deployment", err)
				c.AddError(fmt.Sprintf("Failed to publish deployment on %s!", cluster))
			}
		}
		c.HandleResponse(nil)
		return
	}

	// 拼湊 images 升級
	param.imageMap = make(map[string]string)
	imageArr := strings.Split(param.Images, ",")
	// param.imageMap = make(map[string]string)
	for _, image := range imageArr {
		arr := strings.Split(image, "=")
		if len(arr) == 2 && arr[1] != "" {
			param.imageMap[arr[0]] = arr[1]
		}
	}
	// 拼湊環境變量
	param.envMap = make(map[string]string)
	envArr := strings.Split(param.Environments, ",")
	// param.envMap = make(map[string]string)
	for _, env := range envArr {
		arr := strings.Split(env, "=")
		if len(arr) == 2 && arr[1] != "" {
			param.envMap[arr[0]] = arr[1]
		}
	}

	if len(param.imageMap) == 0 && len(param.envMap) == 0 {
		c.AddErrorAndResponse(fmt.Sprintf("Invalid images/environments parameter: %s %s", param.Images, param.Environments), http.StatusBadRequest)
		return
	}

	deployInfoMap := make(map[int64]([]*DeploymentInfo))
	for _, cluster := range param.clusters {
		deployInfo, err := getOnlineDeploymenetInfo(param.Deployment, param.Namespace, cluster, 0)
		if err != nil {
			c.AddError(fmt.Sprintf("Failed to get online deployment info on %s", cluster))
			continue
		}

		// 率先把強制指定的環境變量，如和系統環境變量衝突，後面會覆蓋
		for k, v := range deployInfo.DeploymentObject.Spec.Template.Spec.Containers {
			for i, e := range v.Env {
				if param.envMap[e.Name] != "" {
					deployInfo.DeploymentObject.Spec.Template.Spec.Containers[k].Env[i].Value = param.envMap[e.Name]
				}
			}
		}

		common.DeploymentPreDeploy(deployInfo.DeploymentObject, deployInfo.Deployment, deployInfo.Cluster, deployInfo.Namespace)
		tmplId := deployInfo.DeploymentTemplete.Id
		deployInfo.DeploymentTemplete.Id = 0
		deployInfo.DeploymentTemplete.User = c.APIKey.String()
		deployInfo.DeploymentTemplete.Description = "[APIKey] " + c.GetString("description")
		// 更新鏡像版本
		ci := make(map[string]string)
		for k, v := range param.imageMap {
			ci[k] = v
		}
		for k, v := range deployInfo.DeploymentObject.Spec.Template.Spec.Containers {
			if param.imageMap[v.Name] != "" {
				deployInfo.DeploymentObject.Spec.Template.Spec.Containers[k].Image = param.imageMap[v.Name]
				delete(ci, v.Name)
			}
		}
		if len(ci) > 0 {
			var keys []string
			for k := range ci {
				keys = append(keys, k)
			}
			c.AddError(fmt.Sprintf("Deployment template don't have container: %s", strings.Join(keys, ",")))
			continue
		}
		deployInfoMap[tmplId] = append(deployInfoMap[tmplId], deployInfo)
	}

	if len(c.Failure.Body.Errors) > 0 {
		c.HandleResponse(nil)
		return
	}

	for id, deployInfos := range deployInfoMap {
		deployInfo := deployInfos[0]
		newTpl, err := json.Marshal(deployInfo.DeploymentObject)
		if err != nil {
			logs.Error("Failed to parse metadata: %s", err)
			c.AddError(fmt.Sprintf("Failed to parse metadata!"))
			continue
		}
		deployInfo.DeploymentTemplete.Template = string(newTpl)
		//更新deploymentTpl中的CreateTime和UpdateTime,數據庫中不會自動更新
		deployInfo.DeploymentTemplete.CreateTime = time.Now()
		deployInfo.DeploymentTemplete.UpdateTime = time.Now()
		newtmplId, err := models.DeploymentTplModel.Add(deployInfo.DeploymentTemplete)
		if err != nil {
			logs.Error("Failed to save new deployment template", err)
			c.AddError(fmt.Sprintf("Failed to save new deployment template!"))
			continue
		}

		for k, deployInfo := range deployInfos {
			err := models.DeploymentModel.UpdateById(deployInfo.Deployment)
			if err != nil {
				logs.Error("Failed to update deployment by id", err)
				c.AddError(fmt.Sprintf("Failed to update deployment by id!"))
				continue
			}
			deployInfoMap[id][k].DeploymentTemplete.Id = newtmplId
		}
	}

	if !param.Publish || len(c.Failure.Body.Errors) > 0 {
		c.HandleResponse(nil)
		return
	}

	for _, deployInfos := range deployInfoMap {
		for _, deployInfo := range deployInfos {
			err := publishDeployment(deployInfo, c.APIKey.String())
			if err != nil {
				logs.Error("Failed to publish deployment", err)
				c.AddError(fmt.Sprintf("Failed to publish deployment on %s!", deployInfo.Cluster.Name))
			}
		}
	}
	c.HandleResponse(nil)
}

// swagger:route GET /scale_deployment deploy ScaleDeploymentParam
//
// 用於 CI/CD 中的部署水平擴容/縮容
//
// 副本數量範圍為0-32
// 該接口只能使用 app 級別的 apikey，這樣做的目的主要是防止 apikey 的濫用
//
//     Responses:
//       200: responseSuccess
//       400: responseState
//       401: responseState
//       500: responseState
// @router /scale_deployment [get]
func (c *OpenAPIController) ScaleDeployment() {
	param := ScaleDeploymentParam{
		Deployment: c.GetString("deployment"),
		Namespace:  c.GetString("namespace"),
		Cluster:    c.GetString("cluster"),
	}
	if !c.CheckoutRoutePermission(ScaleDeploymentAction) || !c.CheckDeploymentPermission(param.Deployment) || !c.CheckNamespacePermission(param.Namespace) {
		return
	}
	var err error
	param.Replicas, err = c.GetInt("replicas", 0)
	if err != nil {
		c.AddErrorAndResponse(fmt.Sprintf("Invalid replicas parameter: %s", err.Error()), http.StatusBadRequest)
		return
	}
	if param.Replicas > 32 || param.Replicas <= 0 {
		c.AddErrorAndResponse(fmt.Sprintf("Invalid replicas parameter: %d not in range (0,32]", param.Replicas), http.StatusBadRequest)
		return
	}
	if len(param.Namespace) == 0 {
		c.AddErrorAndResponse(fmt.Sprintf("Invalid namespace parameter"), http.StatusBadRequest)
		return
	}
	if len(param.Deployment) == 0 {
		c.AddErrorAndResponse(fmt.Sprintf("Invalid deployment parameter"), http.StatusBadRequest)
		return
	}

	ns, err := models.NamespaceModel.GetByName(param.Namespace)
	if err != nil {
		c.AddErrorAndResponse(fmt.Sprintf("Failed get namespace by name(%s)", param.Namespace), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal([]byte(ns.MetaData), &ns.MetaDataObj)
	if err != nil {
		logs.Error(fmt.Sprintf("Failed to parse metadata: %s", err.Error()))
		c.AddErrorAndResponse("", http.StatusInternalServerError)
		return
	}
	deployResource, err := models.DeploymentModel.GetByName(param.Deployment)
	if err != nil {
		c.AddErrorAndResponse(fmt.Sprintf("Failed get deployment by name(%s)", param.Deployment), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal([]byte(deployResource.MetaData), &deployResource.MetaDataObj)
	if err != nil {
		logs.Error(fmt.Sprintf("Failed to parse metadata: %s", err.Error()))
		c.AddErrorAndResponse("", http.StatusInternalServerError)
		return
	}

	cli, err := client.Client(param.Cluster)
	if err != nil {
		logs.Error("Failed to connect to k8s client", err)
		c.AddErrorAndResponse(fmt.Sprintf("Failed to connect to k8s client on %s!", param.Cluster), http.StatusInternalServerError)
		return
	}

	deployObj, err := resdeployment.GetDeployment(cli, param.Deployment, ns.KubeNamespace)
	if err != nil {
		logs.Error("Failed to get deployment from k8s client", err.Error())
		c.AddErrorAndResponse(fmt.Sprintf("Failed to get deployment from k8s client on %s!", param.Cluster), http.StatusInternalServerError)
		return
	}
	replicas32 := int32(param.Replicas)
	deployObj.Spec.Replicas = &replicas32
	if err := updateDeployment(deployObj, param.Cluster, c.APIKey.String(), "Scale Deployment", deployResource.Id); err != nil {
		logs.Error("Failed to upgrade from k8s client", err.Error())
		c.AddErrorAndResponse(fmt.Sprintf("Failed to upgrade from k8s client on %s!", param.Cluster), http.StatusInternalServerError)
		return
	}
	err = models.DeploymentModel.Update(replicas32, deployResource, param.Cluster)
	if err != nil {
		// 非敏感錯誤，無須暴露給用戶
		logs.Error("Failed to update deployment in db!", err.Error())
	}
	c.HandleResponse(nil)
}

// 主要用於從數據庫中查找、拼湊出用於更新的模板資源，資源主要用於 k8s 數據更新和 數據庫存儲更新記錄等
func getOnlineDeploymenetInfo(deployment, namespace, cluster string, templateId int64) (deployInfo *DeploymentInfo, err error) {
	if len(deployment) == 0 {
		return nil, fmt.Errorf("Invalid deployment parameter!")
	}
	if len(namespace) == 0 {
		return nil, fmt.Errorf("Invalid namespace parameter!")
	}
	if len(cluster) == 0 {
		return nil, fmt.Errorf("Invalid cluster parameter!")
	}
	deployResource, err := models.DeploymentModel.GetByName(deployment)
	if err != nil {
		return nil, fmt.Errorf("Failed to get deployment by name(%s)!", deployment)
	}

	deployInfo = new(DeploymentInfo)

	// 根據特定模板升級
	if templateId != 0 {
		deployInfo.DeploymentTemplete, err = models.DeploymentTplModel.GetById(int64(templateId))
		if err != nil {
			return nil, fmt.Errorf("Failed to get deployment template by id: %s", err.Error())
		}
		if deployResource.Id != deployInfo.DeploymentTemplete.DeploymentId {
			return nil, fmt.Errorf("Invalid template id parameter(no permission)!")
		}
	} else {
		// 獲取並更新線上模板
		status, err := models.PublishStatusModel.GetByCluster(models.PublishTypeDeployment, deployResource.Id, cluster)
		if err != nil {
			return nil, fmt.Errorf("Failed to get publish status by cluster: %s", err.Error())
		}
		onlineTplId := status.TemplateId

		deployInfo.DeploymentTemplete, err = models.DeploymentTplModel.GetById(onlineTplId)
		if err != nil {
			return nil, fmt.Errorf("Failed to get deployment template by id: %s", err.Error())
		}
	}

	deployObj := appsv1.Deployment{}
	err = json.Unmarshal(hack.Slice(deployInfo.DeploymentTemplete.Template), &deployObj)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse deployment template: %s", err.Error())
	}
	// 拼湊 namespace 參數
	app, _ := models.AppModel.GetById(deployResource.AppId)
	err = json.Unmarshal([]byte(app.Namespace.MetaData), &app.Namespace.MetaDataObj)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse namespace metadata: %s", err.Error())
	}
	if namespace != app.Namespace.Name {
		return nil, fmt.Errorf("Invalid namespace parameter(should be the namespace of the application)")
	}
	deployObj.Namespace = app.Namespace.KubeNamespace

	// 拼湊副本數量參數
	err = json.Unmarshal([]byte(deployResource.MetaData), &deployResource.MetaDataObj)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse deployment resource metadata: %s", err.Error())
	}
	rp := deployResource.MetaDataObj.Replicas[cluster]
	deployObj.Spec.Replicas = &rp
	deployInfo.DeploymentObject = &deployObj
	deployInfo.Deployment = deployResource

	deployInfo.Namespace = app.Namespace

	deployInfo.Cluster, err = models.ClusterModel.GetParsedMetaDataByName(cluster)
	if err != nil {
		return nil, fmt.Errorf("Failed to get cluster by name: %s", err.Error())
	}
	return deployInfo, nil
}

// 通過給定模板資源把業務發佈到k8s集群中，並在數據庫中更新發布記錄
func publishDeployment(deployInfo *DeploymentInfo, username string) error {
	// 操作 kubernetes api，實現升級部署
	cli, err := client.Client(deployInfo.Cluster.Name)
	if err == nil {
		publishHistory := &models.PublishHistory{
			Type:         models.PublishTypeDeployment,
			ResourceId:   deployInfo.Deployment.Id,
			ResourceName: deployInfo.DeploymentObject.Name,
			TemplateId:   deployInfo.DeploymentTemplete.Id,
			Cluster:      deployInfo.Cluster.Name,
			User:         username,
			Message:      deployInfo.DeploymentTemplete.Description,
		}
		defer models.PublishHistoryModel.Add(publishHistory)
		_, err = resdeployment.CreateOrUpdateDeployment(cli, deployInfo.DeploymentObject)
		if err != nil {
			publishHistory.Status = models.ReleaseFailure
			publishHistory.Message = err.Error()
			return fmt.Errorf("Failed to create or update deployment by k8s client: %s", err.Error())
		} else {
			publishHistory.Status = models.ReleaseSuccess
			err := models.PublishStatusModel.Add(deployInfo.Deployment.Id, deployInfo.DeploymentTemplete.Id, deployInfo.Cluster.Name, models.PublishTypeDeployment)
			if err != nil {
				return err
			}
			return nil
		}
	} else {
		return fmt.Errorf("Failed to get k8s client(cluster: %s): %v", deployInfo.Cluster.Name, err)
	}
}

func updateDeployment(deployObj *appsv1.Deployment, cluster string, name string, msg string, resourceId int64) error {
	status, err := models.PublishStatusModel.GetByCluster(models.PublishTypeDeployment, resourceId, cluster)
	if err != nil {
		return fmt.Errorf("Failed to get publish status by cluster: %s", err.Error())
	}
	publishHistory := &models.PublishHistory{
		Type:         models.PublishTypeDeployment,
		ResourceId:   resourceId,
		ResourceName: deployObj.Name,
		TemplateId:   status.TemplateId,
		Cluster:      cluster,
		User:         name,
		Message:      msg,
	}
	defer models.PublishHistoryModel.Add(publishHistory)
	cli, err := client.Client(cluster)
	if err != nil {
		return err
	}
	_, err = resdeployment.UpdateDeployment(cli, deployObj)
	if err != nil {
		publishHistory.Status = models.ReleaseFailure
		publishHistory.Message = err.Error()
		return fmt.Errorf("Failed to update deployment by k8s client: %s", err.Error())
	} else {
		publishHistory.Status = models.ReleaseSuccess
		err := models.PublishStatusModel.Add(resourceId, status.TemplateId, cluster, models.PublishTypeDeployment)
		if err != nil {
			return err
		}
		return nil
	}
}

// swagger:route GET /get_deployment_detail deploy DeploymentDetailParam
//
// 通過給定的namespace、app name、deployment name來查詢某個具體deployment的信息
//
// 因為查詢範圍是對所有的服務進行的，因此需要綁定 全局 apikey 使用。
//
//     Responses:
//       200: respresourceinfo
//       400: responseState
//       500: responseState
// @router /get_deployment_detail [get]
func (c *OpenAPIController) GetDeploymentDetail() {
	if !c.CheckoutRoutePermission(GetDeploymentDetailAction) {
		return
	}
	if c.APIKey.Type != models.GlobalAPIKey {
		c.AddErrorAndResponse("You can only use global APIKey in this action!", http.StatusUnauthorized)
		return
	}
	ns := c.GetString("namespace")
	app := c.GetString("app")
	deployment := c.GetString("deployment")
	if len(ns) == 0 {
		c.AddErrorAndResponse("Invalid namespace parameter!", http.StatusBadRequest)
		return
	}
	if len(app) == 0 {
		c.AddErrorAndResponse("Invalid app parameter!", http.StatusBadRequest)
		return
	}
	if len(deployment) == 0 {
		c.AddErrorAndResponse("Invalid deployment parameter!", http.StatusBadRequest)
		return
	}
	params := deploymentDetailParam{ns, app, deployment}
	dep, err := models.DeploymentModel.GetUniqueDepByName(params.Namespace, params.App, params.Deployment)
	if err != nil {
		c.AddErrorAndResponse("Failed to get deployment by name!", http.StatusBadRequest)
		return
	}
	resp := new(deploymentDetailResponse)
	resp.Body.Deployment = dep
	resp.Body.Code = http.StatusOK
	c.HandleResponse(resp.Body)
}

// swagger:route GET /get_latest_deployment_tpl deploy LatestDeploymentTplParam
//
// 通過給定的namespace、app name、deployment name來查詢某個具體deployment的最新部署模板信息
//
// 因為查詢範圍是對所有的服務進行的，因此需要綁定 全局 apikey 使用。
//
//     Responses:
//       200: respresourceinfo
//       400: responseState
//       500: responseState
// @router /get_latest_deployment_tpl [get]
func (c *OpenAPIController) GetLatestDeploymentTpl() {
	if !c.CheckoutRoutePermission(GetLatestDeploymentTplAction) {
		return
	}
	if c.APIKey.Type != models.GlobalAPIKey {
		c.AddErrorAndResponse("You can only use global APIKey in this action!", http.StatusUnauthorized)
		return
	}
	ns := c.GetString("namespace")
	app := c.GetString("app")
	deployment := c.GetString("deployment")
	if len(ns) == 0 {
		c.AddErrorAndResponse("Invalid namespace parameter!", http.StatusBadRequest)
		return
	}
	if len(app) == 0 {
		c.AddErrorAndResponse("Invalid app parameter!", http.StatusBadRequest)
		return
	}
	if len(deployment) == 0 {
		c.AddErrorAndResponse("Invalid deployment parameter!", http.StatusBadRequest)
		return
	}
	params := latestDeploymentTplParam{ns, app, deployment}
	dep, err := models.DeploymentTplModel.GetLatestDeptplByName(params.Namespace, params.App, params.Deployment)
	if err != nil {
		c.AddErrorAndResponse("Failed to get deployment by name!", http.StatusBadRequest)
		return
	}
	resp := new(latestDeploymentTplResponse)
	resp.Body.DeploymentTpl = dep
	resp.Body.Code = http.StatusOK
	c.HandleResponse(resp.Body)
}
