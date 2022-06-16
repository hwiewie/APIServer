package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {
	const KubeSecretController = "github.com/hwiewie/APIServer/controllers/kubernetes/secret:KubeSecretController"
	beego.GlobalControllerRouter[KubeSecretController] = append(
		beego.GlobalControllerRouter[KubeSecretController],
		beego.ControllerComments{
			Method:           "Create",
			Router:           `/:secretId/tpls/:tplId/clusters/:cluster`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil,
		})
}
