// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package sale

import (
	"github.com/hexya-addons/account"
	"github.com/hexya-addons/saleTeams"
	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/pool/h"
)

var (
	// GroupSaleLayout enables layouts in sale reports
	GroupSaleLayout *security.Group
	// GroupDeliveryInvoiceAddress enables different delivery and invoice addresses
	GroupDeliveryInvoiceAddress *security.Group
	// GroupShowPriceSubtotal shows line subtotals without taxes (B2B)
	GroupShowPriceSubtotal *security.Group
	// GroupShowPriceTotal shows line subtotals with taxes (B2C)
	GroupShowPriceTotal *security.Group
	// GroupMRPProperties shows properties on sale lines
	GroupMRPProperties *security.Group
	// GroupDiscountPerSOLine shows discount on each sale lines
	GroupDiscountPerSOLine *security.Group
	// GroupDisplayIncoterms shows incoterms in sale orders
	GroupDisplayIncoterms *security.Group
	// GroupWarningSale allows warnings to be set on a product or a customer (sale)
	GroupWarningSale *security.Group
	// GroupAnalyticAccounting enables analytic accounting for sales
	GroupAnalyticAccounting *security.Group
)

func init() {
	GroupSaleLayout = security.Registry.NewGroup("sale_group_sale_layout", "Personalize sale order and invoice report")
	GroupDeliveryInvoiceAddress = security.Registry.NewGroup("sale_group_delivery_invoice_address", "Addresses in Sales Orders")
	GroupShowPriceSubtotal = security.Registry.NewGroup("sale_group_show_price_subtotal", "Show line subtotals without taxes (B2B)")
	GroupShowPriceTotal = security.Registry.NewGroup("sale_group_show_price_total", "Show line subtotals with taxes included (B2C)")
	GroupMRPProperties = security.Registry.NewGroup("sale_group_mrp_properties", "Properties on lines")
	GroupDiscountPerSOLine = security.Registry.NewGroup("sale_group_discount_per_so_line", "Discount on lines")
	GroupDisplayIncoterms = security.Registry.NewGroup("sale_group_display_incoterm", "Display incoterms on Sales Order and related invoices")
	GroupWarningSale = security.Registry.NewGroup("sale_group_warning_sale", "A warning can be set on a product or a customer (Sale)")
	GroupAnalyticAccounting = security.Registry.NewGroup("sale_group_analytic_accounting", "Analytic Accounting for Sales")

	h.SaleOrder().Methods().Load().AllowGroup(saleTeams.GroupSaleSalesman)
	h.SaleOrder().Methods().Write().AllowGroup(saleTeams.GroupSaleSalesman)
	h.SaleOrder().Methods().Create().AllowGroup(saleTeams.GroupSaleSalesman)
	h.SaleOrderLine().Methods().AllowAllToGroup(saleTeams.GroupSaleSalesman)
	h.SaleOrderLine().Methods().Read().AllowGroup(account.GroupAccountUser)
	h.SaleOrderLine().Methods().Write().AllowGroup(account.GroupAccountUser)
	h.AccountInvoiceTax().Methods().AllowAllToGroup(saleTeams.GroupSaleSalesman)
	h.AccountInvoice().Methods().AllowAllToGroup(saleTeams.GroupSaleSalesman)
	h.AccountInvoice().Methods().AllowAllToGroup(saleTeams.GroupSaleManager)
	h.AccountInvoiceLine().Methods().AllowAllToGroup(saleTeams.GroupSaleSalesman)
	h.AccountPaymentTerm().Methods().Load().AllowGroup(saleTeams.GroupSaleSalesman)
	h.AccountAnalyticTag().Methods().Load().AllowGroup(saleTeams.GroupSaleSalesman)
	h.AccountAnalyticAccount().Methods().AllowAllToGroup(saleTeams.GroupSaleSalesman)
	h.SaleOrder().Methods().AllowAllToGroup(saleTeams.GroupSaleManager)
	h.SaleOrder().Methods().AllowAllToGroup(account.GroupAccountUser)
	h.SaleReport().Methods().AllowAllToGroup(saleTeams.GroupSaleSalesman)
	h.SaleReport().Methods().AllowAllToGroup(saleTeams.GroupSaleManager)
	h.AccountJournal().Methods().Load().AllowGroup(saleTeams.GroupSaleSalesman)
	h.Partner().Methods().Load().AllowGroup(saleTeams.GroupSaleSalesman)
	h.Partner().Methods().AllowAllToGroup(saleTeams.GroupSaleManager)
	h.ProductTemplate().Methods().Load().AllowGroup(saleTeams.GroupSaleSalesman)
	h.ProductProduct().Methods().Load().AllowGroup(saleTeams.GroupSaleSalesman)
	h.AccountTax().Methods().Load().AllowGroup(saleTeams.GroupSaleSalesman)
	h.Attachment().Methods().AllowAllToGroup(saleTeams.GroupSaleSalesman)
	h.Attachment().Methods().AllowAllToGroup(saleTeams.GroupSaleManager)
	h.ProductUom().Methods().Load().AllowGroup(saleTeams.GroupSaleSalesman)
	h.ProductPricelist().Methods().Load().AllowGroup(saleTeams.GroupSaleSalesman)
	h.AccountAccount().Methods().Load().AllowGroup(saleTeams.GroupSaleSalesman)
	h.ProductUomCategory().Methods().AllowAllToGroup(saleTeams.GroupSaleManager)
	h.ProductUom().Methods().AllowAllToGroup(saleTeams.GroupSaleManager)
	h.ProductCategory().Methods().AllowAllToGroup(saleTeams.GroupSaleManager)
	h.ProductSupplierinfo().Methods().Load().AllowGroup(saleTeams.GroupSaleSalesman)
	h.ProductSupplierinfo().Methods().AllowAllToGroup(saleTeams.GroupSaleManager)
	h.ProductPricelist().Methods().AllowAllToGroup(saleTeams.GroupSaleManager)
	h.Partner().Methods().AllowAllToGroup(saleTeams.GroupSaleManager)
	h.AccountMoveLine().Methods().Load().AllowGroup(saleTeams.GroupSaleSalesman)
	h.SaleOrder().Methods().AllowAllToGroup(account.GroupAccountInvoice)
	h.SaleOrderLine().Methods().AllowAllToGroup(account.GroupAccountInvoice)
	h.SaleLayoutCategory().Methods().AllowAllToGroup(saleTeams.GroupSaleManager)
	h.SaleLayoutCategory().Methods().AllowAllToGroup(account.GroupAccountManager)
	h.SaleLayoutCategory().Methods().AllowAllToGroup(saleTeams.GroupSaleSalesman)
	h.SaleLayoutCategory().Methods().AllowAllToGroup(saleTeams.GroupSaleSalesmanAllLeads)
	h.SaleLayoutCategory().Methods().Load().AllowGroup(account.GroupAccountInvoice)
	h.ProductPricelistItem().Methods().AllowAllToGroup(saleTeams.GroupSaleManager)
	h.ProductPriceHistory().Methods().AllowAllToGroup(saleTeams.GroupSaleManager)
	h.ProductTemplate().Methods().AllowAllToGroup(saleTeams.GroupSaleManager)
	h.ProductProduct().Methods().AllowAllToGroup(saleTeams.GroupSaleManager)
	h.ProductAttribute().Methods().AllowAllToGroup(saleTeams.GroupSaleManager)
	h.ProductAttributeValue().Methods().AllowAllToGroup(saleTeams.GroupSaleManager)
	h.ProductAttributePrice().Methods().AllowAllToGroup(saleTeams.GroupSaleManager)
	h.ProductAttributeLine().Methods().AllowAllToGroup(saleTeams.GroupSaleManager)
	h.AccountTax().Methods().Load().AllowGroup(saleTeams.GroupSaleSalesman)
	h.AccountJournal().Methods().Load().AllowGroup(saleTeams.GroupSaleSalesman)
	h.AccountInvoiceTax().Methods().Load().AllowGroup(saleTeams.GroupSaleSalesman)
	h.AccountTaxGroup().Methods().Load().AllowGroup(saleTeams.GroupSaleSalesman)
	h.AccountAccount().Methods().Load().AllowGroup(saleTeams.GroupSaleManager)

}
