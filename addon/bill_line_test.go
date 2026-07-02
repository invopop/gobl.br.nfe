package addon_test

import (
	"testing"

	"github.com/invopop/gobl.br.nfe/addon"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/num"
	"github.com/invopop/gobl/org"
	"github.com/invopop/gobl/regimes/br"
	"github.com/invopop/gobl/rules"
	"github.com/invopop/gobl/tax"
	"github.com/stretchr/testify/assert"
)

func TestLineValidation(t *testing.T) {
	tests := []struct {
		name string
		line *bill.Line
		err  string
	}{
		{
			name: "valid line with all required taxes",
			line: &bill.Line{
				Index:    1,
				Quantity: num.MakeAmount(1, 0),
				Sum:      num.NewAmount(100, 2),
				Total:    num.NewAmount(100, 2),
				Item: &org.Item{
					Name:  "Test",
					Price: num.NewAmount(100, 2),
				},
				Taxes: tax.Set{
					{
						Category: br.TaxCategoryICMS,
						Ext: tax.ExtensionsOf(cbc.CodeMap{
							addon.ExtKeyICMSCST:    "00",
							addon.ExtKeyICMSOrigin: "0",
						}),
					},
					{
						Category: br.TaxCategoryPIS,
						Ext:      tax.ExtensionsOf(cbc.CodeMap{addon.ExtKeyPISCST: "01"}),
					},
					{
						Category: br.TaxCategoryCOFINS,
						Ext:      tax.ExtensionsOf(cbc.CodeMap{addon.ExtKeyCOFINSCST: "01"}),
					},
				},
			},
		},
		{
			name: "nil line",
			line: nil,
		},
		{
			name: "missing taxes",
			line: &bill.Line{},
			err:  "line taxes must include the ICMS category",
		},
		{
			name: "empty taxes",
			line: &bill.Line{
				Taxes: tax.Set{},
			},
			err: "line taxes must include the ICMS category",
		},
		{
			name: "missing ICMS tax",
			line: &bill.Line{
				Taxes: tax.Set{
					{
						Category: br.TaxCategoryPIS,
					},
					{
						Category: br.TaxCategoryCOFINS,
					},
				},
			},
			err: "line taxes must include the ICMS category",
		},
		{
			name: "missing PIS tax",
			line: &bill.Line{
				Taxes: tax.Set{
					{
						Category: br.TaxCategoryICMS,
					},
					{
						Category: br.TaxCategoryCOFINS,
					},
				},
			},
			err: "line taxes must include the PIS category",
		},
		{
			name: "missing COFINS tax",
			line: &bill.Line{
				Taxes: tax.Set{
					{
						Category: br.TaxCategoryICMS,
					},
					{
						Category: br.TaxCategoryPIS,
					},
				},
			},
			err: "line taxes must include the COFINS category",
		},
	}

	for _, ts := range tests {
		t.Run(ts.name, func(t *testing.T) {
			err := rules.Validate(ts.line, withAddonContext())
			if ts.err == "" {
				assert.NoError(t, err)
			} else {
				if assert.Error(t, err) {
					assert.Contains(t, err.Error(), ts.err)
				}
			}
		})
	}
}

func withAddonContext() rules.WithContext {
	return func(rc *rules.Context) {
		rc.Set(rules.ContextKey(addon.V4), tax.AddonForKey(addon.V4))
	}
}
