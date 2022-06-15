package openapi

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/labels"

	"github.com/hwiewie/APIServer/client"
	"github.com/hwiewie/APIServer/client/api"
	"github.com/hwiewie/APIServer/models"
	"github.com/hwiewie/APIServer/models/response"
	"github.com/hwiewie/APIServer/resources/common"
	"github.com/hwiewie/APIServer/resources/pod"
	"github.com/hwiewie/APIServer/util/logs"
)

// An array of the pod.
// swagger:response resppodlist
type resppodlist struct {
	// in: body
	// Required: true
	Body struct {
		response.ResponseBase
		Pods []response.Pod `json:"pods"`
	}
}

// swagger:parameters PodInfoParam
type PodInfoParam struct {
	// in: query
	// Pod Label Key,只允許填寫一個
	// Required: true
	LabelSelector string `json:"labelSelector"`
	// Required: true
	Cluster string `json:"cluster"`
}

// swagger:parameters PodInfoFromIPParam
type PodInfoFromIPParam struct {
	// in: query
	// Pod IP 列表，使用逗號分隔
	// Required: true
	IPS string `json:"ips"`
	ips map[string]bool
	// Required: true
	Cluster string `json:"cluster"`
}

// swagger:parameters PodListParam
type PodListParam struct {
	// Wayne 的 namespace 名稱，必須與 Name 同時存在或者不存在
	// in: query
	// Required: false
	Namespace string `json:"namespace"`
	// 資源名稱，必須與 Namespace 同時存在或者不存在
	// in: query
	// Required: false
	Name string `json:"name"`
	// 資源類型:daemonsets,deployments,cronjobs,statefulsets
	// in: query
	// Required: true
	Type api.ResourceName `json:"type"`
}

// An array of the pod.
// swagger:response respPodInfoList
type respPodInfoList struct {
	// in: body
	// Required: true
	Body struct {
		response.ResponseBase
		RespListInfo []*respListInfo `json:"list"`
	}
}

type respListInfo struct {
	Cluster      string `json:"cluster,omitempty"`
	ResourceName string `json:"resourceName,omitempty"`
	// Wayne namespace 名稱
	Namespace string        `json:"namespace,omitempty"`
	Pods      []respPodInfo `json:"pods"`
}

type respPodInfo struct {
	Name      string    `json:"name,omitempty"`
	Namespace string    `json:"namespace,omitempty"`
	NodeName  string    `json:"nodeName,omitempty"`
	PodIp     string    `json:"podIp,omitempty"`
	State     string    `json:"state,omitempty"`
	StartTime time.Time `json:"startTime,omitempty"`
}

// swagger:route GET /get_pod_info pod PodInfoParam
//
// 用於獲取線上所有 pod 中包含請求條件中 labelSelector 指定的特定 label 的 pod
//
// 返回 每個 pod 的 pod IP 和 所有 label 列表。
// 需要綁定全局 apikey 使用。該接口的權限控制為只能使用全局 apikey 的原因是查詢條件為 labelSelector ，是對所有 app 的 條件過濾。
//
//     Responses:
//       200: resppodlist
//       401: responseState
//       500: responseState
// @router /get_pod_info [get]
func (c *OpenAPIController) GetPodInfo() {
	if !c.CheckoutRoutePermission(GetPodInfoAction) {
		return
	}
	if c.APIKey.Type != models.GlobalAPIKey {
		c.AddErrorAndResponse("You can only use global APIKey in this action!", http.StatusUnauthorized)
		return
	}
	podList := resppodlist{}
	podList.Body.Code = http.StatusOK
	params := PodInfoParam{c.GetString("labelSelector"), c.GetString("cluster")}
	if params.Cluster == "" {
		c.AddErrorAndResponse("Invalid cluster parameter:must required!", http.StatusBadRequest)
		return
	}
	manager, err := client.Manager(params.Cluster)
	if err != nil {
		c.AddErrorAndResponse("Invalid cluster parameter:not exist!", http.StatusBadRequest)
		return
	}

	pods, err := pod.ListPodByLabelKey(manager.CacheFactory, "", params.LabelSelector)
	if err != nil {
		logs.Error(fmt.Sprintf("Failed to parse metadata: %s", err.Error()))
		c.AddErrorAndResponse(fmt.Sprintf("Maybe a problematic k8s cluster(%s)!", params.Cluster), http.StatusInternalServerError)
		return
	}
	for _, p := range pods {
		podList.Body.Pods = append(podList.Body.Pods, response.Pod{Labels: p.Labels, PodIp: p.PodIp})
	}
	c.HandleResponse(podList.Body)
}

// swagger:route GET /get_pod_info_from_ip pod PodInfoFromIPParam
//
// 用於通過線上 kubernetes Pod IP 反查對應 Pod 信息的接口
//
// 返回 每個 pod 的 pod IP 和 所有 label 列表。
// 需要綁定全局 apikey 使用。該接口的權限控制為只能使用全局 apikey 的原因是查詢條件為 IP ，是對所有 app 的 條件過濾。
//
//     Responses:
//       200: resppodlist
//       401: responseState
//       500: responseState
// @router /get_pod_info_from_ip [get]
func (c *OpenAPIController) GetPodInfoFromIP() {
	if !c.CheckoutRoutePermission(GetPodInfoFromIPAction) {
		return
	}
	if c.APIKey.Type != models.GlobalAPIKey {
		c.AddErrorAndResponse("You can only use global APIKey in this action!", http.StatusUnauthorized)
		return
	}
	params := PodInfoFromIPParam{IPS: c.GetString("ips"), Cluster: c.GetString("cluster")}
	if params.Cluster == "" {
		c.AddErrorAndResponse("Invalid cluster parameter:must required!", http.StatusBadRequest)
		return
	}
	params.ips = make(map[string]bool)
	for _, ip := range strings.Split(params.IPS, ",") {
		params.ips[ip] = true
	}
	manager, err := client.Manager(params.Cluster)
	if err != nil {
		c.AddErrorAndResponse("Invalid cluster parameter:not exist!", http.StatusBadRequest)
		return
	}
	pods, err := pod.ListKubePod(manager.CacheFactory, "", nil)
	if err != nil {
		logs.Error(fmt.Sprintf("Failed to parse metadata: %s", err.Error()))
		c.AddErrorAndResponse(fmt.Sprintf("Maybe a problematic k8s cluster(%s)!", params.Cluster), http.StatusInternalServerError)
		return
	}
	podList := resppodlist{}
	podList.Body.Code = http.StatusOK
	for _, p := range pods {
		if params.ips[p.Status.PodIP] {
			podList.Body.Pods = append(podList.Body.Pods, response.Pod{Labels: p.Labels, PodIp: p.Status.PodIP})
		}
	}
	c.HandleResponse(podList.Body)

}

// swagger:route GET /get_pod_list pod PodListParam
//
// 用於根據資源類型獲取所有機房Pod列表
//
// 返回 Pod 信息
// 需要綁定全局 apikey 使用。
//
//     Responses:
//       200: respPodInfoList
//       401: responseState
//       500: responseState
// @router /get_pod_list [get]
func (c *OpenAPIController) GetPodList() {
	if !c.CheckoutRoutePermission(GetPodListAction) {
		return
	}
	if c.APIKey.Type != models.GlobalAPIKey {
		c.AddErrorAndResponse("You can only use global APIKey in this action!", http.StatusUnauthorized)
		return
	}
	podList := respPodInfoList{}
	podList.Body.Code = http.StatusOK
	params := PodListParam{c.GetString("namespace"), c.GetString("name"), c.GetString("type")}
	if params.Type == "" {
		c.AddErrorAndResponse("Invalid type parameter:must required!", http.StatusBadRequest)
		return
	}
	if (params.Name == "" && params.Namespace != "") || (params.Name != "" && params.Namespace == "") {
		c.AddErrorAndResponse("Namespace and Name must exist or not exist at the same time!", http.StatusBadRequest)
		return
	}
	var ns *models.Namespace
	var err error
	if params.Namespace != "" {
		ns, err = models.NamespaceModel.GetByName(params.Namespace)
		if err != nil {
			c.AddErrorAndResponse(fmt.Sprintf("Get Namespace by name (%s) error!%v", params.Namespace, err), http.StatusBadRequest)
			return
		}
	}

	managers := client.Managers()
	managers.Range(func(key, value interface{}) bool {
		manager := value.(*client.ClusterManager)
		// if Name and Namespace empty,return all pods
		if params.Name == "" && params.Namespace == "" {
			objs, err := manager.KubeClient.List(params.Type, "", labels.Everything().String())
			if err != nil {
				c.AddError(fmt.Sprintf("List all resource error.cluster:%s,type:%s, %v",
					manager.Cluster.Name, params.Type, err))
				return true
			}

			for _, obj := range objs {
				commonObj, err := common.ToBaseObject(obj)
				if err != nil {
					c.AddError(fmt.Sprintf("ToBaseObject error.cluster:%s,type:%s, %v",
						manager.Cluster.Name, params.Type, err))
					return true
				}
				podListResp, err := buildPodListResp(manager, params.Namespace, commonObj.Namespace, commonObj.Name, params.Type)
				if err != nil {
					c.AddError(fmt.Sprintf("buildPodListResp error.cluster:%s,type:%s, %v",
						manager.Cluster.Name, params.Type, err))
					return true
				}
				if len(podListResp.Pods) > 0 {
					podList.Body.RespListInfo = append(podList.Body.RespListInfo, podListResp)
				}
			}
			return true
		}

		podListResp, err := buildPodListResp(manager, params.Namespace, ns.KubeNamespace, params.Name, params.Type)
		if err != nil {
			c.AddError(fmt.Sprintf("buildPodListResp error.cluster:%s,type:%s,namespace:%s,name:%s %v",
				manager.Cluster.Name, ns.KubeNamespace, params.Name, params.Type, err))
			return true
		}
		if len(podListResp.Pods) > 0 {
			podList.Body.RespListInfo = append(podList.Body.RespListInfo, podListResp)
		}

		return true
	})

	c.HandleResponse(podList.Body)
}

func buildPodListResp(manager *client.ClusterManager, namespace, kubeNamespace, resourceName string, resourceType api.ResourceName) (*respListInfo, error) {
	pods, err := pod.GetPodListByType(manager.KubeClient, kubeNamespace, resourceName, resourceType)
	if err != nil {
		return nil, err
	}

	listInfo := &respListInfo{
		Cluster:      manager.Cluster.Name,
		ResourceName: resourceName,
		Namespace:    namespace,
	}

	for _, kubePod := range pods {
		listInfo.Pods = append(listInfo.Pods, respPodInfo{
			Name:      kubePod.Name,
			Namespace: kubePod.Namespace,
			NodeName:  kubePod.Spec.NodeName,
			PodIp:     kubePod.Status.PodIP,
			State:     pod.GetPodStatus(kubePod),
			StartTime: kubePod.CreationTimestamp.Time,
		})
	}
	return listInfo, nil
}
