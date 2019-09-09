// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package sale

import (
	_ "github.com/hexya-addons/account"
	_ "github.com/hexya-addons/procurement"
	_ "github.com/hexya-addons/saleTeams"
	"github.com/hexya-addons/web/controllers"
	"github.com/hexya-erp/hexya/src/server"
)

const MODULE_NAME = "sale"

func init() {
	server.RegisterModule(&server.Module{
		Name:     MODULE_NAME,
		PostInit: func() {},
	})

	controllers.BackendCSS = append(controllers.BackendCSS, "/static/sale/src/css/sale.css")
	controllers.BackendJS = append(controllers.BackendJS, "/static/sale/src/js/sale.js")
}
