package initial

import (
	_ "github.com/hwiewie/APIServer/controllers/auth/db"
	_ "github.com/hwiewie/APIServer/controllers/auth/ldap"
	_ "github.com/hwiewie/APIServer/controllers/auth/oauth2"
	_ "github.com/hwiewie/APIServer/oauth2"
)
