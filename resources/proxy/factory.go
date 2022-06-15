package proxy

import (
	"encoding/json"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/hwiewie/APIServer/client/api"
	"github.com/hwiewie/APIServer/resources/dataselector"
)

func getRealObjCellByKind(name api.ResourceName, object runtime.Object) (dataselector.DataCell, error) {
	switch name {
	// There are special filtering requests to add resource processing here
	case api.ResourceNamePod:
		obj := object.(*v1.Pod)
		return PodCell(*obj), nil
	default:
		objByte, err := json.Marshal(object)
		if err != nil {
			return nil, err
		}
		var commonObj ObjectCell
		err = json.Unmarshal(objByte, &commonObj)
		if err != nil {
			return nil, err
		}
		return commonObj, nil
	}
}
