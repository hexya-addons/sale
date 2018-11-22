// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package sale

import (
	_ "github.com/hexya-addons/account"
	_ "github.com/hexya-addons/procurement"
	_ "github.com/hexya-addons/saleTeams"
	"github.com/hexya-addons/web/controllers"
	"github.com/hexya-erp/hexya/src/models/security"
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

	GroupSaleLayout = security.Registry.NewGroup("sale_group_sale_layout", "Personalize sale order and invoice report")
	GroupDeliveryInvoiceAddress = security.Registry.NewGroup("sale_group_delivery_invoice_address", "Addresses in Sales Orders")
	GroupShowPriceSubtotal = security.Registry.NewGroup("sale_group_show_price_subtotal", "Show line subtotals without taxes (B2B)")
	GroupShowPriceTotal = security.Registry.NewGroup("sale_group_show_price_total", "Show line subtotals with taxes included (B2C)")
}
