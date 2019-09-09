// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package sale

import (
	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/pool/h"
	"github.com/hexya-erp/pool/m"
	"github.com/hexya-erp/pool/q"
)

func init() {

	h.ProductProduct().AddFields(map[string]models.FieldDefinition{
		"SalesCount": models.IntegerField{String: "# Sales", Compute: h.ProductProduct().Methods().ComputeSalesCount(),
			GoType: new(int)},
	})

	h.ProductProduct().Methods().ComputeSalesCount().DeclareMethod(
		`ComputeSalesCount returns the number of sales for this product`,
		func(rs m.ProductProductSet) m.ProductProductData {
			cond := q.SaleReport().State().In([]string{"sale", "done"}).And().Product().In(rs)
			return h.ProductProduct().NewData().SetSalesCount(
				h.SaleReport().NewSet(rs.Env()).Search(cond).SearchCount())
		})
}
