// Copyright 2019 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package sale

import (
	"testing"

	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/hexya/src/models/types/dates"
	"github.com/hexya-erp/hexya/src/tests"
	"github.com/hexya-erp/pool/h"
	"github.com/hexya-erp/pool/q"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMain(m *testing.M) {
	tests.RunTests(m, "sale", nil)
}

func TestOnChangeProduct(t *testing.T) {
	Convey("Test Onchange Product", t, func() {
		So(models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			Convey("Testing Product Onchange", func() {
				uom := h.ProductUom().Search(env, q.ProductUom().Name().Equals("Unit(s)"))
				priceList := h.ProductPricelist().Search(env, q.ProductPricelist().Name().Equals("Public Pricelist"))
				partner := h.Partner().Create(env, h.Partner().NewData().SetName("George"))
				taxInclude := h.AccountTax().Create(env, h.AccountTax().NewData().
					SetName("Include tax").
					SetAmount(21).
					SetPriceInclude(true).
					SetTypeTaxUse("sale"))
				taxExclude := h.AccountTax().Create(env, h.AccountTax().NewData().
					SetName("Exclude tax").
					SetAmount(0).
					SetTypeTaxUse("sale"))
				productTmpl := h.ProductTemplate().Create(env, h.ProductTemplate().NewData().
					SetName("Voiture").
					SetListPrice(121).
					SetTaxes(taxInclude).
					SetUom(uom).
					SetUomPo(uom))
				product := h.ProductProduct().Create(env, h.ProductProduct().NewData().SetProductTmpl(productTmpl))
				fp := h.AccountFiscalPosition().Create(env, h.AccountFiscalPosition().NewData().
					SetName("fiscal position").
					SetSequence(1))
				h.AccountFiscalPositionTax().Create(env, h.AccountFiscalPositionTax().NewData().
					SetPosition(fp).
					SetTaxSrc(taxInclude).
					SetTaxDest(taxExclude))
				order := h.SaleOrder().Create(env, h.SaleOrder().NewData().
					SetPartner(partner).
					SetPricelist(priceList).
					SetFiscalPosition(fp).
					CreateOrderLine(
						h.SaleOrderLine().NewData().
							SetName(product.Name()).
							SetProduct(product).
							SetProductUomQty(1).
							SetProductUom(uom).
							SetPriceUnit(121)))
				soLine := order.OrderLine().Records()[0]
				res := soLine.ProductChange()
				So(res.HasPriceUnit(), ShouldBeTrue)
				So(res.PriceUnit(), ShouldEqual, 100)
			})
			Convey("Test different price lists are correctly applied based on dates", func() {
				uom := h.ProductUom().Search(env, q.ProductUom().Name().Equals("Unit(s)"))
				supportProduct := h.ProductProduct().NewSet(env).GetRecord("product_product_product_2")
				supportProduct.SetListPrice(100)
				partner := h.Partner().Create(env, h.Partner().NewData().SetName("George"))

				christmasPricelist := h.ProductPricelist().Create(env, h.ProductPricelist().NewData().
					SetName("Christmas pricelist").
					CreateItems(h.ProductPricelistItem().NewData().
						SetDateStart(dates.ParseDate("2017-12-01")).
						SetDateEnd(dates.ParseDate("2017-12-24")).
						SetComputePrice("percentage").
						SetBase("ListPrice").
						SetPercentPrice(20).
						SetAppliedOn("3_global")).
					CreateItems(h.ProductPricelistItem().NewData().
						SetDateStart(dates.ParseDate("2017-12-25")).
						SetDateEnd(dates.ParseDate("2017-12-31")).
						SetComputePrice("percentage").
						SetBase("ListPrice").
						SetPercentPrice(50).
						SetAppliedOn("3_global")))
				so := h.SaleOrder().Create(env, h.SaleOrder().NewData().
					SetPartner(partner).
					SetDateOrder(dates.ParseDateTime("2017-12-20 12:00:00")).
					SetPricelist(christmasPricelist))
				orderLine := h.SaleOrderLine().Create(env, h.SaleOrderLine().NewData().
					SetName("Dummy").
					SetProductUomQty(1).
					SetProductUom(uom).
					SetOrder(so).
					SetProduct(supportProduct))

				// force compute uom and prices
				orderLine.Write(orderLine.ProductChange())
				orderLine.Write(orderLine.ProductUomChange())
				So(orderLine.PriceUnit(), ShouldEqual, 80)

				so.SetDateOrder(dates.ParseDateTime("2017-12-30 12:00:00"))
				orderLine.Write(orderLine.ProductChange())
				So(orderLine.PriceUnit(), ShouldEqual, 50)
			})
			Convey("Test prices and discounts are correctly applied based on date and uom", func() {
				uom := h.ProductUom().Search(env, q.ProductUom().Name().Equals("Unit(s)"))
				computerCase := h.ProductProduct().NewSet(env).GetRecord("product_product_product_16")
				computerCase.SetListPrice(100)
				partner := h.Partner().Create(env, h.Partner().NewData().SetName("George"))
				categUnit := h.ProductUomCategory().NewSet(env).GetRecord("product_product_uom_categ_unit")
				groupDiscount := h.Group().NewSet(env).Search(q.Group().GroupID().Equals(GroupDiscountPerSOLine.ID))
				currentUser := h.User().NewSet(env).CurrentUser()
				currentUser.SetGroups(currentUser.Groups().Union(groupDiscount))
				currentUser.SyncMemberships()
				newUom := h.ProductUom().Create(env, h.ProductUom().NewData().
					SetName("10 Units").
					SetFactorInv(10).
					SetUomType("bigger").
					SetRounding(1.0).
					SetCategory(categUnit))
				christmasPricelist := h.ProductPricelist().Create(env, h.ProductPricelist().NewData().
					SetName("Christmas pricelist").
					SetDiscountPolicy("without_discount").
					CreateItems(h.ProductPricelistItem().NewData().
						SetDateStart(dates.ParseDate("2017-12-01")).
						SetDateEnd(dates.ParseDate("2017-12-30")).
						SetComputePrice("percentage").
						SetBase("ListPrice").
						SetPercentPrice(10).
						SetAppliedOn("3_global")))
				so := h.SaleOrder().Create(env, h.SaleOrder().NewData().
					SetPartner(partner).
					SetDateOrder(dates.ParseDateTime("2017-12-20 12:00:00")).
					SetPricelist(christmasPricelist))
				orderLine := h.SaleOrderLine().Create(env, h.SaleOrderLine().NewData().
					SetName("Dummy").
					SetProductUomQty(1).
					SetProductUom(uom).
					SetOrder(so).
					SetProduct(computerCase))

				// force compute uom and prices
				orderLine.Write(orderLine.ProductChange())
				orderLine.Write(orderLine.ProductUomChange())
				orderLine.Write(orderLine.OnchangeDiscount())
				So(orderLine.PriceSubtotal(), ShouldEqual, 90)
				So(orderLine.Discount(), ShouldEqual, 10)

				orderLine.SetProductUom(newUom)
				orderLine.Write(orderLine.ProductUomChange())
				orderLine.Write(orderLine.OnchangeDiscount())
				So(orderLine.PriceSubtotal(), ShouldEqual, 900)
				So(orderLine.Discount(), ShouldEqual, 10)
			})
		}), ShouldBeNil)
	})
}
