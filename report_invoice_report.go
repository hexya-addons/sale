// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package sale

import (
	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/pool/h"
	"github.com/hexya-erp/pool/m"
)

func init() {

	h.AccountInvoiceReport().AddFields(map[string]models.FieldDefinition{
		"Team": models.Many2OneField{String: "Sales Team", RelationModel: h.CRMTeam()},
	})

	h.AccountInvoiceReport().Methods().Select().Extend("",
		func(rs m.AccountInvoiceReportSet) string {
			return rs.Super().Select() + ", sub.team_id as team_id"
		})

	h.AccountInvoiceReport().Methods().SubSelect().Extend("",
		func(rs m.AccountInvoiceReportSet) string {
			return rs.Super().SubSelect() + ", ai.team_id as team_id"
		})

	h.AccountInvoiceReport().Methods().GroupByClause().Extend("",
		func(rs m.AccountInvoiceReportSet) string {
			return rs.Super().GroupByClause() + ", ai.team_id"
		})

}
