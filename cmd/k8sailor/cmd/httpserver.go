package cmd

import (
	"github.com/hwiewie/APIServer/cmd/k8sailor/global"
	"github.com/hwiewie/APIServer/internal/apis"
	"github.com/hwiewie/APIServer/internal/k8scache"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var cmdHttpserver = &cobra.Command{
	Use:  "httpserver",
	Long: "啟動 web 服務器",
	Run: func(cmd *cobra.Command, args []string) {
		// 啟動 informer
		runInformer()

		// 啟動服務
		runHttpserver()
	},
}

// runHttpserver 啟動 http server
func runHttpserver() {
	// 1. 將 apis 註冊到 httpserver 中
	global.HttpServer.RegisterRoute(apis.RootGroup)

	// 2. 啟動服務
	if err := global.HttpServer.Run(); err != nil {
		logrus.Fatalf("start httpserver failed: %v", err)
	}
}

func runInformer() {

	clientset := global.KubeClient.Client()
	informer := global.KubeInformer.WithClientset(clientset)

	k8scache.RegisterHandlers(informer)

	informer.Start()
}
