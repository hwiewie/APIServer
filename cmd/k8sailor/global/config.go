package global

import (
	"github.com/go-jarvis/jarvis"
	"github.com/hwiewie/APIServer/pkg/confgin"
	"github.com/hwiewie/APIServer/pkg/confk8s"
)

// 定義服務相關信息
var (
	HttpServer   = &confgin.Server{}
	KubeClient   = &confk8s.Client{}
	KubeInformer = &confk8s.Informer{}

	app = jarvis.App{
		Name: "APIServer",
	}
)

// 使用 jarvis 初始化配置文件
func init() {

	config := &struct {
		HttpServer   *confgin.Server
		KubeClient   *confk8s.Client
		KubeInformer *confk8s.Informer
	}{
		HttpServer:   HttpServer,
		KubeClient:   KubeClient,
		KubeInformer: KubeInformer,
	}

	app.Conf(config)
}
