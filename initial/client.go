package initial

import (
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/hwiewie/APIServer/client"
)

func InitClient() {
	// 定期更新client
	go wait.Forever(client.BuildApiserverClient, 5*time.Second)
}
