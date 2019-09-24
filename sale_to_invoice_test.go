// Copyright 2019 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package sale

import (
	"testing"

	"github.com/hexya-addons/account"
	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/hexya/src/models/types"
	"github.com/hexya-erp/hexya/src/models/types/dates"
	"github.com/hexya-erp/pool/h"
	"github.com/hexya-erp/pool/q"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSale(t *testing.T) {
	Convey("Testing Sale To Invoice", t, func() {
		So(models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			Convey("Testing for invoice create,validate and pay with invoicing and payment user.", func() {
				group := h.Group().Search(env, q.Group().GroupID().Equals(account.GroupAccountInvoice.ID))
				company := h.Company().NewSet(env).GetRecord("base_main_company")
				userTypeRevenue := h.AccountAccountType().NewSet(env).GetRecord("account_data_account_type_revenue")
				userTypeReceivable := h.AccountAccountType().NewSet(env).GetRecord("account_data_account_type_receivable")
				accountRev := h.AccountAccount().Create(env, h.AccountAccount().NewData().
					SetCode("X2020").
					SetName("Sales - Test Sales Account").
					SetUserType(userTypeRevenue).
					SetReconcile(true))
				accountRecv := h.AccountAccount().Create(env, h.AccountAccount().NewData().
					SetCode("X1012").
					SetName("Sales - Test Reicv Account").
					SetUserType(userTypeReceivable).
					SetReconcile(true))
				// Add account to product
				productTemplate := h.ProductProduct().NewSet(env).GetRecord("sale_advance_product_0").ProductTmpl()
				productTemplate.SetPropertyAccountIncome(accountRev)
				// Create sale journal
				h.AccountJournal().Create(env, h.AccountJournal().NewData().
					SetName("Sale Journal - Test").
					SetCode("STSJ").
					SetType("sale").
					SetCompany(company))
				// In order to test, I create new user and applied Invoicing & Payments group.
				user := h.User().Create(env, h.User().NewData().
					SetName("Test User").
					SetLogin("test@test.com").
					SetCompany(company).
					SetGroups(group))
				user.SyncMemberships()
				// I create partner for sale order.
				partner := h.Partner().Create(env, h.Partner().NewData().
					SetName("Test Customer").
					SetEmail("testcustomer@test.com").
					SetPropertyAccountReceivable(accountRecv))
				// In order to test I create sale order and confirmed it.
				order := h.SaleOrder().Create(env, h.SaleOrder().NewData().
					SetPartner(partner).
					SetPartnerInvoice(partner).
					SetPartnerShipping(partner).
					SetDateOrder(dates.Now()).
					SetPricelist(h.ProductPricelist().NewSet(env).GetRecord("product_list0")))
				ctx := types.NewContext().
					WithKey("active_model", "SaleOrder").
					WithKey("active_id", order.ID()).
					WithKey("active_ids", order.Ids())
				order.WithNewContext(ctx).ActionConfirm()
				// Now I create invoice.
				payment := h.SaleAdvancePaymentInv().Create(env, h.SaleAdvancePaymentInv().NewData().
					SetAdvancePaymentMethod("fixed").
					SetAmount(5).
					SetProduct(h.ProductProduct().NewSet(env).GetRecord("sale_advance_product_0")))
				payment.WithNewContext(ctx).CreateInvoices()
				So(order.Invoices().IsNotEmpty(), ShouldBeTrue)
				for _, invoice := range order.Invoices().Records() {
					invoice.WithNewContext(ctx).InvoiceValidate()
				}
			})
		}), ShouldBeNil)
	})
}
