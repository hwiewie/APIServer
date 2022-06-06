package common

import (
	"io/ioutil"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func InitClient() (clientset *kubernetes.Clientset, err error) {
	var (
		restConf *rest.Config
	)

	if restConf, err = GetRestConf(); err != nil {
		return
	}

	if clientset, err = kubernetes.NewForConfig(restConf); err != nil {
		goto END
	}
END:
	return
}

func GetRestConf() (restConf *rest.Config, err error) {
	var (
		kubeconfig []byte
	)

	if kubeconfig, err = ioutil.ReadFile("./admin.conf"); err != nil {
		goto END
	}

	if restConf, err = clientcmd.RESTConfigFromKubeConfig(kubeconfig); err != nil {
		goto END
	}
END:
	return
}
