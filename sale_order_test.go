// Copyright 2019 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package sale

import (
	"testing"

	"github.com/hexya-addons/saleTeams"
	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/pool/h"
	"github.com/hexya-erp/pool/m"
	"github.com/hexya-erp/pool/q"
	. "github.com/smartystreets/goconvey/convey"
)

type saleOrderTestData struct {
	Manager         m.UserSet
	User            m.UserSet
	ProductOrder    m.ProductProductSet
	ProductDelivery m.ProductProductSet
	ServiceOrder    m.ProductProductSet
	ServiceDelivery m.ProductProductSet
	Partner         m.PartnerSet
	SaleJournal     m.AccountJournalSet
}

func initTestSaleOrder(env models.Environment) *saleOrderTestData {
	var res saleOrderTestData
	groupManager := h.Group().Search(env, q.Group().GroupID().Equals(saleTeams.GroupSaleManager.ID))
	groupUser := h.Group().Search(env, q.Group().GroupID().Equals(saleTeams.GroupSaleSalesman.ID))
	res.Manager = h.User().Create(env, h.User().NewData().
		SetName("Andrew Manager").
		SetLogin("manager").
		SetEmail("a.m@example.com").
		SetSignature("--\nAndrew").
		//SetNotifyEmail("always").
		SetGroups(groupManager))
	res.Manager.SyncMemberships()
	res.User = h.User().Create(env, h.User().NewData().
		SetName("Mark Manager").
		SetLogin("user").
		SetEmail("m.u@example.com").
		SetSignature("--\nMark").
		//SetNotifyEmail("always").
		SetGroups(groupUser))
	res.User.SyncMemberships()
	res.ProductOrder = h.ProductProduct().NewSet(env).GetRecord("product_product_order_01")
	res.ProductDelivery = h.ProductProduct().NewSet(env).GetRecord("product_product_delivery_01")
	res.ServiceOrder = h.ProductProduct().NewSet(env).GetRecord("product_service_order_01")
	res.ServiceDelivery = h.ProductProduct().NewSet(env).GetRecord("product_service_delivery")
	res.Partner = h.Partner().NewSet(env).GetRecord("base_res_partner_1")
	res.SaleJournal = h.AccountJournal().Create(env, h.AccountJournal().NewData().
		SetName("Sale's journal").
		SetType("sale").
		SetCode("SJ"))
	return &res
}

func TestSaleOrder(t *testing.T) {
	Convey("Testing sale orders", t, func() {
		So(models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			chart := h.AccountChartTemplate().NewSet(env).GetRecord("l10n_generic_coa_configurable_chart_template")
			chart.TryLoadingForCurrentCompany()
			tsd := initTestSaleOrder(env)
			Convey(`Test the sale order flow (invoicing and quantity updates) 
- Invoice repeatedly while varrying delivered quantities and check that invoice are always what we expect`, func() {
				soData := h.SaleOrder().NewData().
					SetPartner(tsd.Partner).
					SetPartnerInvoice(tsd.Partner).
					SetPartnerShipping(tsd.Partner).
					SetPricelist(h.ProductPricelist().NewSet(env).GetRecord("product_list0"))
				var amountTotal, amountOrder float64
				for _, p := range []m.ProductProductSet{tsd.ProductOrder, tsd.ProductDelivery, tsd.ServiceOrder, tsd.ServiceDelivery} {
					soData.CreateOrderLine(h.SaleOrderLine().NewData().
						SetName(p.Name()).
						SetProduct(p).
						SetProductUomQty(2).
						SetProductUom(p.Uom()).
						SetPriceUnit(p.ListPrice()))
					amountTotal += 2 * p.ListPrice()
					if p.InvoicePolicy() == "order" {
						amountOrder += 2 * p.ListPrice()
					}
				}
				so := h.SaleOrder().Create(env, soData)
				So(so.AmountTotal(), ShouldEqual, amountTotal)
				So(so.Name(), ShouldNotBeEmpty)
				So(so.Name(), ShouldNotEqual, "False")
				So(so.Name(), ShouldNotEqual, "New")

				// Send quotation
				so.ForceQuotationSend()
				So(so.State(), ShouldEqual, "sent")

				// Confirm quotation
				so.ActionConfirm()
				So(so.State(), ShouldEqual, "sale")
				So(so.InvoiceStatus(), ShouldEqual, "to invoice")

				// create invoice: only 'invoice on order' products are invoiced
				inv := so.ActionInvoiceCreate(false, false)
				So(inv.InvoiceLines().Len(), ShouldEqual, 2)
				So(inv.AmountTotal(), ShouldEqual, amountOrder)
				So(so.InvoiceStatus(), ShouldEqual, "no")
				So(so.Invoices().Len(), ShouldEqual, 1)
			})
			Convey("Test deleting and cancelling sale orders depending on their state and on the user's rights", func() {
				soData := h.SaleOrder().NewData().
					SetPartner(tsd.Partner).
					SetPartnerShipping(tsd.Partner).
					SetPartnerInvoice(tsd.Partner).
					SetPricelist(h.ProductPricelist().NewSet(env).GetRecord("product_list0"))
				for _, p := range []m.ProductProductSet{tsd.ProductOrder, tsd.ProductDelivery, tsd.ServiceOrder, tsd.ServiceDelivery} {
					soData.CreateOrderLine(h.SaleOrderLine().NewData().
						SetName(p.Name()).
						SetProduct(p).
						SetProductUomQty(2).
						SetProductUom(p.Uom()).
						SetPriceUnit(p.ListPrice()))
				}
				so := h.SaleOrder().Create(env, soData)

				// SO in state 'draft' can be deleted
				soCopy := so.Copy(nil)

				So(soCopy.State(), ShouldEqual, "draft")
				So(func() { soCopy.Sudo(tsd.User.ID()).Unlink() }, ShouldPanic)
				So(soCopy.State(), ShouldEqual, "draft")
				soCopy.Sudo(tsd.Manager.ID()).Unlink()

				// SO in state 'cancel' can be deleted
				soCopy = so.Copy(nil)
				soCopy.ActionConfirm()
				So(soCopy.State(), ShouldEqual, "sale")
				soCopy.ActionCancel()
				So(soCopy.State(), ShouldEqual, "cancel")
				So(func() { soCopy.Sudo(tsd.User.ID()).Unlink() }, ShouldPanic)
				So(func() { soCopy.Sudo(tsd.Manager.ID()).Unlink() }, ShouldNotPanic)

				// SO in state 'sale' or 'done' cannot be deleted
				so.ActionConfirm()
				So(so.State(), ShouldEqual, "sale")
				So(func() { so.Sudo(tsd.Manager.ID()).Unlink() }, ShouldPanic)

				so.ActionDone()
				So(so.State(), ShouldEqual, "done")
				So(func() { so.Sudo(tsd.Manager.ID()).Unlink() }, ShouldPanic)
			})
			Convey("Test confirming a vendor invoice to reinvoice cost on the so", func() {
				list0 := h.ProductPricelist().NewSet(env).GetRecord("product_list0")
				company := h.Company().NewSet(env).GetRecord("base_main_company")
				list0.SetCurrency(company.Currency())
				servCost := h.ProductProduct().NewSet(env).GetRecord("product_service_cost_01")
				prodGap := h.ProductProduct().NewSet(env).GetRecord("product_product_product_1")
				so := h.SaleOrder().Create(env, h.SaleOrder().NewData().
					SetPartner(tsd.Partner).
					SetPartnerInvoice(tsd.Partner).
					SetPartnerShipping(tsd.Partner).
					SetPricelist(list0).
					CreateOrderLine(h.SaleOrderLine().NewData().
						SetName(prodGap.Name()).
						SetProduct(prodGap).
						SetProductUomQty(2).
						SetProductUom(prodGap.Uom()).
						SetPriceUnit(prodGap.ListPrice())))
				so.ActionConfirm()
				so.CreateAnalyticAccount("")
				invPartner := h.Partner().NewSet(env).GetRecord("base_res_partner_2")
				journal := h.AccountJournal().Create(env, h.AccountJournal().NewData().
					SetName("Purchase Journal").
					SetCode("STPJ").
					SetType("purchase").
					SetCompany(company))
				accountPayable := h.AccountAccount().Create(env, h.AccountAccount().NewData().
					SetCode("X1111").
					SetName("Sale - Test Payable Account").
					SetUserType(h.AccountAccountType().NewSet(env).GetRecord("account_data_account_type_payable")).
					SetReconcile(true))
				accountIncome := h.AccountAccount().Create(env, h.AccountAccount().NewData().
					SetCode("X1112").
					SetName("Sale - Test Account").
					SetUserType(h.AccountAccountType().NewSet(env).GetRecord("account_data_account_type_direct_costs")))
				invoiceVals := h.AccountInvoice().NewData().
					SetName("").
					SetType("in_invoice").
					SetPartner(invPartner).
					SetAccount(accountPayable).
					SetJournal(journal).
					SetCurrency(company.Currency()).
					CreateInvoiceLines(h.AccountInvoiceLine().NewData().
						SetName(servCost.Name()).
						SetProduct(servCost).
						SetQuantity(2).
						SetUom(servCost.Uom()).
						SetPriceUnit(servCost.StandardPrice()).
						SetAccountAnalytic(so.Project()).
						SetAccount(accountIncome))
				inv := h.AccountInvoice().Create(env, invoiceVals)
				inv.ActionInvoiceOpen()
				sol := so.OrderLine().Filtered(func(r m.SaleOrderLineSet) bool {
					return r.Product().Equals(servCost)
				})
				So(sol.IsNotEmpty(), ShouldBeTrue)
				So(sol.PriceUnit(), ShouldEqual, 160)
				So(sol.QtyDelivered(), ShouldEqual, 2)
				So(sol.ProductUomQty(), ShouldEqual, 0)
				So(sol.QtyInvoiced(), ShouldEqual, 0)
			})
		}), ShouldBeNil)
	})
}
