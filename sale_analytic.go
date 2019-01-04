// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package sale

import (
	"math"

	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/pool/h"
	"github.com/hexya-erp/pool/q"
)

func init() {

	h.SaleOrderLine().Methods().ComputeAnalytic().DeclareMethod(
		`ComputeAnalytic updates analytic lines linked with this SaleOrderLine`,
		func(rs h.SaleOrderLineSet, cond q.AccountAnalyticLineCondition) bool {
			lines := make(map[int64]float64)
			forceSOLines := rs.Env().Context().GetIntegerSlice("force_so_lines")
			if cond.IsEmpty() {
				if rs.IsEmpty() && len(forceSOLines) == 0 {
					return true
				}
				cond = q.AccountAnalyticLine().SoLine().In(rs).And().Amount().LowerOrEqual(0)
			}
			data := h.AccountAnalyticLine().Search(rs.Env(), cond).
				GroupBy(q.AccountAnalyticLine().ProductUom(), q.AccountAnalyticLine().SoLine()).
				Aggregates(q.AccountAnalyticLine().ProductUom(), q.AccountAnalyticLine().SoLine(),
					q.AccountAnalyticLine().UnitAmount())
			for _, d := range data {
				uom := d.Values.ProductUom()
				line := d.Values.SoLine()
				qty := d.Values.UnitAmount()
				if line.ProductUom().Category().Equals(uom.Category()) {
					qty = uom.ComputeQuantity(qty, line.ProductUom(), true)
				}
				lines[line.ID()] += qty
			}
			for l, qty := range lines {
				h.SaleOrderLine().Browse(rs.Env(), []int64{l}).SetQtyDelivered(qty)
			}
			return true
		})

	h.AccountAnalyticLine().AddFields(map[string]models.FieldDefinition{
		"SoLine": models.Many2OneField{String: "Sale Order Line", RelationModel: h.SaleOrderLine()},
	})

	h.AccountAnalyticLine().Methods().GetInvoicePrice().DeclareMethod(
		`GetInvoicePrice returns the unit price to set on invoice`,
		func(rs h.AccountAnalyticLineSet, order h.SaleOrderSet) float64 {
			if rs.Product().ExpensePolicy() == "sales_price" {
				return rs.Product().
					WithContext("partner", order.Partner().ID()).
					WithContext("date_order", order.DateOrder()).
					WithContext("pricelist", order.Pricelist().ID()).
					WithContext("uom", rs.ProductUom().ID()).Price()
			}
			if rs.UnitAmount() == 0 {
				return 0
			}
			// Prevent unnecessary currency conversion that could be impacted by exchange rate
			// fluctuations
			if !rs.Currency().IsEmpty() && rs.AmountCurrency() != 0 && rs.Currency().Equals(order.Currency()) {
				return math.Abs(rs.AmountCurrency() / rs.UnitAmount())
			}
			priceUnit := math.Abs(rs.Amount() / rs.UnitAmount())
			currency := rs.Company().Currency()
			if !currency.IsEmpty() && !currency.Equals(order.Currency()) {
				priceUnit = currency.Compute(priceUnit, order.Currency(), true)
			}
			return priceUnit
		})

	h.AccountAnalyticLine().Methods().GetSaleOrderLineVals().DeclareMethod(
		`GetSaleOrderLineVals returns the data to create a sale order line from this account analytic line on
		the given order for the given price.`,
		func(rs h.AccountAnalyticLineSet, order h.SaleOrderSet, price float64) *h.SaleOrderLineData {
			lastSOLine := h.SaleOrderLine().Search(rs.Env(), q.SaleOrderLine().Order().Equals(order)).
				OrderBy("Sequence DESC").Limit(1)
			lastSequence := int64(100)
			if !lastSOLine.IsEmpty() {
				lastSequence = lastSOLine.Sequence() + 1
			}
			fPos := order.Partner().PropertyAccountPosition()
			if !order.FiscalPosition().IsEmpty() {
				fPos = order.FiscalPosition()
			}
			taxes := fPos.MapTax(rs.Product().Taxes(), rs.Product(), order.Partner())

			return h.SaleOrderLine().NewData().
				SetOrder(order).
				SetName(rs.Name()).
				SetSequence(lastSequence).
				SetPriceUnit(price).
				SetTax(taxes).
				SetDiscount(0).
				SetProduct(rs.Product()).
				SetProductUom(rs.ProductUom()).
				SetProductUomQty(0).
				SetQtyDelivered(rs.UnitAmount())
		})

	h.AccountAnalyticLine().Methods().GetSaleOrderLine().DeclareMethod(
		`GetSaleOrderLine adds the sale order line data to the given vals.
		Returned data is a modified copy of vals.`,
		func(rs h.AccountAnalyticLineSet, vals *h.AccountAnalyticLineData) *h.AccountAnalyticLineData {
			result := *vals
			SOLine := result.SoLine()
			if SOLine.IsEmpty() {
				SOLine = rs.SoLine()
			}
			if !SOLine.IsEmpty() || rs.Account().IsEmpty() || rs.Product().IsEmpty() || rs.Product().ExpensePolicy() == "no" {
				return &result
			}
			orderInSale := h.SaleOrder().Search(rs.Env(),
				q.SaleOrder().Project().Equals(rs.Account()).
					And().State().Equals("sale")).Limit(1)
			order := orderInSale
			if order.IsEmpty() {
				order = h.SaleOrder().Search(rs.Env(), q.SaleOrder().Project().Equals(rs.Account())).Limit(1)
			}
			if order.IsEmpty() {
				return &result
			}
			price := rs.GetInvoicePrice(order)
			SOLines := h.SaleOrderLine().Search(rs.Env(),
				q.SaleOrderLine().Order().Equals(order).
					And().PriceUnit().Equals(price).
					And().Product().Equals(rs.Product()))
			if !SOLines.IsEmpty() {
				result.SetSoLine(SOLines.Records()[0])
				return &result
			}
			if order.State() != "sale" {
				panic(rs.T("The Sale Order %s linked to the Analytic Account must be validated before registering expenses.", order.Name()))
			}
			orderLineVals := rs.GetSaleOrderLineVals(order, price)
			NewSOLine := h.SaleOrderLine().Create(rs.Env(), orderLineVals)

			NewSOLine.Write(NewSOLine.ComputeTax())
			result.SetSoLine(NewSOLine)

			return &result
		})

	h.AccountAnalyticLine().Methods().Write().Extend("",
		func(rs h.AccountAnalyticLineSet, data *h.AccountAnalyticLineData) bool {
			if rs.Env().Context().GetBool("create") {
				return rs.Super().Write(data)
			}
			res := rs.Super().Write(data)
			for _, line := range rs.Records() {
				vals := line.Sudo().GetSaleOrderLine(data)
				rs.Super().Write(vals)
			}
			SOLines := h.SaleOrderLine().NewSet(rs.Env())
			for _, line := range rs.Records() {
				SOLines = SOLines.Union(line.SoLine())
			}
			SOLines.ComputeAnalytic(q.AccountAnalyticLineCondition{})
			return res
		})

	h.AccountAnalyticLine().Methods().Create().Extend("",
		func(rs h.AccountAnalyticLineSet, data *h.AccountAnalyticLineData) h.AccountAnalyticLineSet {
			line := rs.Super().Create(data)
			vals := line.Sudo().GetSaleOrderLine(data)
			line.WithContext("create", true).Write(vals)
			SOLines := h.SaleOrderLine().NewSet(rs.Env())
			for _, l := range rs.Records() {
				SOLines = SOLines.Union(l.SoLine())
			}
			SOLines.ComputeAnalytic(q.AccountAnalyticLineCondition{})
			return line
		})

	h.AccountAnalyticLine().Methods().Unlink().Extend("",
		func(rs h.AccountAnalyticLineSet) int64 {
			SOLines := h.SaleOrderLine().NewSet(rs.Env())
			for _, line := range rs.Records() {
				SOLines = SOLines.Union(line.SoLine())
			}
			res := rs.Super().Unlink()
			SOLines.WithContext("force_so_lines", SOLines.Ids()).ComputeAnalytic(q.AccountAnalyticLineCondition{})
			return res
		})

}
