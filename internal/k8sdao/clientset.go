package k8sdao

import "github.com/hwiewie/APIServer/cmd/k8sailor/global"

var clientset = global.KubeClient.Client()
