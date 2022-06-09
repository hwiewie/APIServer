package k8scache

import (
	"github.com/hwiewie/APIServer/pkg/confk8s"
)

// Handler Group
var (
	DepTank   = &DeploymentCache{}
	EventTank = &PodEventCache{}
)

func RegisterHandlers(informer *confk8s.Informer) {
	informer.WithEventHandlers(DepTank, EventTank)
}