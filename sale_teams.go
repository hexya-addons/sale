// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package sale

import (
	"time"

	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/types/dates"
	"github.com/hexya-erp/hexya/src/tools/nbutils"
	"github.com/hexya-erp/pool/h"
	"github.com/hexya-erp/pool/q"
)

func init() {

	h.CRMTeam().AddFields(map[string]models.FieldDefinition{
		"UseQuotations": models.BooleanField{String: "Quotations", Default: models.DefaultValue(true),
			OnChange: h.CRMTeam().Methods().OnchangeUseQuotation(),
			Help:     "Check this box to manage quotations in this sales team."},
		"UseInvoices": models.BooleanField{String: "Invoices",
			Help: "Check this box to manage invoices in this sales team."},
		"Invoiced": models.FloatField{String: "Invoiced This Month",
			Compute: h.CRMTeam().Methods().ComputeInvoiced(),
			Help: `Invoice revenue for the current month. This is the amount the sales
team has invoiced this month. It is used to compute the progression ratio
of the current and target revenue on the kanban view.`},
		"InvoicedTarget": models.FloatField{String: "Invoice Target",
			Help: `Target of invoice revenue for the current month. This is the amount the sales
team estimates to be able to invoice this month.`},
		"SalesToInvoiceAmount": models.FloatField{String: "Amount of sales to invoice",
			Compute: h.CRMTeam().Methods().ComputeSalesToInvoiceAmount()},
		"Currency": models.Many2OneField{RelationModel: h.Currency(), Related: "Company.Currency", ReadOnly: true,
			Required: true},
	})

	h.CRMTeam().Methods().ComputeSalesToInvoiceAmount().DeclareMethod(
		`ComputeSalesToInvoiceAmount computes the total amount of sale orders that have not yet been invoiced`,
		func(rs h.CRMTeamSet) *h.CRMTeamData {
			amounts := h.SaleOrder().Search(rs.Env(),
				q.SaleOrder().Team().Equals(rs).
					And().InvoiceStatus().Equals("to invoice")).
				GroupBy(q.SaleOrder().Team()).
				Aggregates(q.SaleOrder().Team(), q.SaleOrder().AmountTotal())
			if len(amounts) == 0 {
				return h.CRMTeam().NewData()
			}
			amount := amounts[0].Values.AmountTotal()
			return h.CRMTeam().NewData().SetSalesToInvoiceAmount(amount)
		})

	h.CRMTeam().Methods().ComputeInvoiced().DeclareMethod(
		`ComputeInvoiced returns the total amount invoiced by this sale team this month.`,
		func(rs h.CRMTeamSet) *h.CRMTeamData {
			firstDayOfMonth := dates.Date{Time: time.Date(dates.Today().Year(), dates.Today().Month(), 1,
				0, 0, 0, 0, time.UTC)}
			invoices := h.AccountInvoice().Search(rs.Env(),
				q.AccountInvoice().State().In([]string{"open", "paid"}).
					And().Team().Equals(rs).
					And().Date().LowerOrEqual(dates.Today()).
					And().Date().GreaterOrEqual(firstDayOfMonth).
					And().Type().In([]string{"out_invoice", "out_refund"})).
				GroupBy(q.AccountInvoice().Team()).
				Aggregates(q.AccountInvoice().Team(), q.AccountInvoice().AmountUntaxedSigned())
			if len(invoices) == 0 {
				return h.CRMTeam().NewData()
			}
			amount := invoices[0].Values.AmountUntaxedSigned()
			return h.CRMTeam().NewData().SetInvoiced(amount)
		})

	h.CRMTeam().Methods().UpdateInvoicedTarget().DeclareMethod(
		`UpdateInvoicedTarget updates the invoice target with the given value`,
		func(rs h.CRMTeamSet, value float64) bool {
			return rs.Write(h.CRMTeam().NewData().SetInvoicedTarget(nbutils.Round(value, 1)))
		})

	h.CRMTeam().Methods().OnchangeUseQuotation().DeclareMethod(
		`OnchangeUseQuotation makes sure we use invoices if we use quotations.`,
		func(rs h.CRMTeamSet) *h.CRMTeamData {
			return h.CRMTeam().NewData().SetUseInvoices(rs.UseQuotations())
		})

}
