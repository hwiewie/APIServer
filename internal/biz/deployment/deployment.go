package deployment

import (
	"context"
	"fmt"

	"github.com/hwiewie/APIServer/internal/biz/pod"
	"github.com/hwiewie/APIServer/internal/biz/replicaset"
	"github.com/hwiewie/APIServer/internal/k8scache"
	"github.com/hwiewie/APIServer/internal/k8sdao"
	appsv1 "k8s.io/api/apps/v1"
)

type Deployment struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`

	// Replicas 實際期望的 pod 數量
	Replicas int32 `json:"replicas"`

	// 鏡像列表
	Images []string `json:"images"`

	Status DeploymentStatus `json:"status"`

	Labels map[string]string `json:"labelSelector"`
}

type DeploymentStatus struct {
	// 標籤匹配的 Pod 數量
	Replicas int32 `json:"replicas"`
	// 可用 pod 數量
	AvailableReplicas int32 `json:"availableReplicas"`
	// 不可用數量
	UnavailableReplicas int32 `json:"unavailableReplicas"`
}

type ListDeploymentsInput struct {
	Namespace string `query:"namespace"`
}

// ListDeployments 獲取 namespace 下的所有 deployments
// 業務層，可以對接不同來源的數據。
func ListDeployments(ctx context.Context, input ListDeploymentsInput) ([]*Deployment, error) {

	/* k8s api 返回的數據 */
	// v1Deps, err := k8sdao.ListDeployments(ctx, input.Namespace)

	/* 使用 informer 保存在本地的 cache 數據 */
	v1Deps, err := k8scache.DepTank.ListDeployments(ctx, input.Namespace)

	if err != nil {
		return nil, err
	}

	deps := make([]*Deployment, len(v1Deps))
	for i, item := range v1Deps {
		deps[i] = extractDeployment(item)
	}

	return deps, nil
}

type GetDeploymentByNameInput struct {
	Namespace string `query:"namespace"`
	Name      string `uri:"name"`
}

// GetDeploymentByName 通過名稱獲取 deployment
func GetDeploymentByName(ctx context.Context, input GetDeploymentByNameInput) (*Deployment, error) {

	/* k8s api 返回的數據 */
	// v1dep, err := k8sdao.GetDeploymentByName(ctx, input.Namespace, input.Name)

	/* 使用本地的 k8scache */
	v1dep, err := k8scache.DepTank.GetDeploymentByName(ctx, input.Namespace, input.Name)
	if err != nil {
		return nil, err
	}

	dep := extractDeployment(*v1dep)
	return dep, nil
}

type GetPodsByDeploymentInput struct {
	Namespace string `query:"namespace"`
	Name      string `uri:"name"`
}

// GetPodsByDeployment 根據 Deployment name 獲取所有 Pod
func GetPodsByDeployment(ctx context.Context, input GetPodsByDeploymentInput) ([]*pod.Pod, error) {
	// get deployment
	dInput := GetDeploymentByNameInput{
		Namespace: input.Namespace,
		Name:      input.Name,
	}
	dep, err := GetDeploymentByName(ctx, dInput)
	if err != nil {
		return nil, err
	}

	// get active replica set
	rsInput := replicaset.ListReplicaSetInput{
		Namespace: dep.Namespace,
		Labels:    dep.Labels,
	}
	rsList, err := replicaset.ListReplicaSet(ctx, rsInput)
	if err != nil {
		return nil, err
	}

	// get pods
	allPods := []*pod.Pod{}
	for _, rs := range rsList {
		pInput := pod.GetPodsByLabelsInput{
			Namespace: rs.Namespace,
			Labels:    rs.Labels,
		}

		pods, err := pod.GetPodsByLabels(ctx, pInput)
		if err != nil {
			return nil, err
		}
		// fmt.Println(len(pods), pods)
		allPods = append(allPods, pods...)
	}

	return allPods, nil
}

// SetDeploymentReplicasInput 調整 deployment pod 數量參數
// Replicas 為了避免 **0值** 影響。
//   1. 使用為 *int 指針對象， 自行在業務邏輯中進行校驗
//   2. 另外也可以使用， `binding` tag， 由 gin 框架的 valicator 幫忙校驗。 https://github.com/go-playground/validator
// Namespace 設置了默認值， 如果請求不提供將由 gin 框架自己填充。
type SetDeploymentReplicasInput struct {
	Namespace string `query:"namespace,default=default"`
	Name      string `uri:"name"`
	Replicas  *int   `query:"replicas" binding:"required"`
}

// SetDeploymentReplicas 設置 deployment 的 pod 副本數量
func SetDeploymentReplicas(ctx context.Context, input SetDeploymentReplicasInput) (bool, error) {

	// 參數驗證
	if input.Replicas == nil {
		err := fmt.Errorf("replicas must be provide")
		return false, err
	}

	err := k8sdao.SetDeploymentReplicas(ctx, input.Namespace, input.Name, *input.Replicas)

	// err==nil -> true 表示設置成功
	// 這裡選擇不在 api 層進行判斷是為了保持各層的業務分工一致性。
	return err == nil, err
}

type DeleteDeploymentByNameInput struct {
	Name      string `uri:"name"`
	Namespace string `query:"namespace"`
}

// DeleteDeploymentByName 根據名字刪除 deployment
func DeleteDeploymentByName(ctx context.Context, input DeleteDeploymentByNameInput) error {
	err := k8sdao.DeleteDeploymentByName(ctx, input.Namespace, input.Name)
	if err != nil {
		return fmt.Errorf("k8s internal error: %w", err)
	}
	return nil
}

type CreateDeploymentByNameInput struct {
	Namespace string `query:"namespace"`
	Name      string `uri:"name"`
	Body      struct {
		Replicas   *int32             `json:"replicas"`
		Containers []k8sdao.Container `json:"containers"`
	} `body:"" mime:"json"`
}

func CreateDeploymentByName(ctx context.Context, input CreateDeploymentByNameInput) (*Deployment, error) {
	depInput := k8sdao.CreateDeploymentInput{
		Name:       input.Name,
		Replicas:   input.Body.Replicas,
		Containers: input.Body.Containers,
	}

	v1dep, err := k8sdao.CreateDeployment(ctx, input.Namespace, depInput)
	if err != nil {
		return nil, err
	}

	dep := extractDeployment(*v1dep)
	return dep, nil
}

// extractDeployment 轉換成業務本身的 Deployment
func extractDeployment(item appsv1.Deployment) *Deployment {
	return &Deployment{
		Name:      item.Name,
		Namespace: item.Namespace,
		Replicas:  *item.Spec.Replicas,
		Images:    pod.PodImages(item.Spec.Template.Spec),
		Status: DeploymentStatus{
			Replicas:            item.Status.Replicas,
			AvailableReplicas:   item.Status.AvailableReplicas,
			UnavailableReplicas: item.Status.UnavailableReplicas,
		},
		Labels: item.Spec.Selector.MatchLabels,
	}
}
