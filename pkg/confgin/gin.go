package confgin

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// Server 定義一個 gin httpserver 的關鍵字段
// `env:""` 可以通過 `github.com/go-jarvis/jarvis` 庫渲染成配置文件
type Server struct {
	Host    string `env:""`
	Port    int    `env:""`
	Appname string `env:""`
	engine  *gin.Engine
}

// SetDefaults 設置默認值
func (s *Server) SetDefaults() {
	if s.Port == 0 {
		s.Port = 80
	}
	if s.Appname == "" {
		s.Appname = "app"
	}
}

// Init 初始化關鍵信息
func (s *Server) Init() {

	if s.engine == nil {
		s.SetDefaults()
		s.engine = gin.Default()
	}

}

// Run 啟動業務
func (s *Server) Run() error {
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)

	return s.engine.Run(addr)
}

// RegisterRoute 註冊
func (s *Server) RegisterRoute(registerFunc func(rg *gin.RouterGroup)) {

	// 將跨域設置到 base 上， 使用 put 請求的時候 OPTIONS 產生 404
	s.engine.Use(MiddleCors())

	// 註冊以服務名為根的路由信息，方便在 k8s ingress 中做轉發
	base := s.engine.Group(s.Appname)

	// 針對 appname 下的路由，允許跨域
	// base.Use(MiddleCors())

	// 註冊業務子路由
	registerFunc(base)
}

func AppendGroup(base *gin.RouterGroup, register func(base *gin.RouterGroup)) {
	register(base)
}