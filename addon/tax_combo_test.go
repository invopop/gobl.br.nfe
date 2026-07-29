package addon_test

import (
	"testing"

	"github.com/invopop/gobl.br.nfe/addon"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/norm"
	"github.com/invopop/gobl/num"
	"github.com/invopop/gobl/org"
	"github.com/invopop/gobl/regimes/br"
	"github.com/invopop/gobl/rules"
	"github.com/invopop/gobl/tax"
	"github.com/stretchr/testify/assert"
)

func TestTaxComboValidation(t *testing.T) {
	tests := []struct {
		name string
		tc   *tax.Combo
		err  string
	}{
		{
			name: "valid ICMS with CST",
			tc: &tax.Combo{
				Category: br.TaxCategoryICMS,
				Ext: tax.ExtensionsOf(cbc.CodeMap{
					addon.ExtKeyICMSCST:    "00",
					addon.ExtKeyICMSOrigin: "0",
				}),
			},
		},
		{
			name: "valid ICMS with CSOSN",
			tc: &tax.Combo{
				Category: br.TaxCategoryICMS,
				Ext: tax.ExtensionsOf(cbc.CodeMap{
					addon.ExtKeyICMSCSOSN:  "102",
					addon.ExtKeyICMSOrigin: "0",
				}),
			},
		},
		{
			name: "ICMS missing situation code",
			tc: &tax.Combo{
				Category: br.TaxCategoryICMS,
				Ext: tax.ExtensionsOf(cbc.CodeMap{
					addon.ExtKeyICMSOrigin: "0",
				}),
			},
			err: "ICMS tax combo requires 'br-nfe-icms-cst' or 'br-nfe-icms-csosn' extension",
		},
		{
			name: "ICMS missing origin",
			tc: &tax.Combo{
				Category: br.TaxCategoryICMS,
				Ext: tax.ExtensionsOf(cbc.CodeMap{
					addon.ExtKeyICMSCST: "00",
				}),
			},
			err: "ICMS tax combo requires 'br-nfe-icms-origin' extension",
		},
		{
			name: "valid PIS",
			tc: &tax.Combo{
				Category: br.TaxCategoryPIS,
				Ext: tax.ExtensionsOf(cbc.CodeMap{
					addon.ExtKeyPISCST: "01",
				}),
			},
		},
		{
			name: "PIS missing CST",
			tc: &tax.Combo{
				Category: br.TaxCategoryPIS,
			},
			err: "PIS tax combo requires 'br-nfe-pis-cst' extension",
		},
		{
			name: "valid COFINS",
			tc: &tax.Combo{
				Category: br.TaxCategoryCOFINS,
				Ext: tax.ExtensionsOf(cbc.CodeMap{
					addon.ExtKeyCOFINSCST: "01",
				}),
			},
		},
		{
			name: "COFINS missing CST",
			tc: &tax.Combo{
				Category: br.TaxCategoryCOFINS,
			},
			err: "COFINS tax combo requires 'br-nfe-cofins-cst' extension",
		},
		{
			name: "ICMS with both CST and CSOSN",
			tc: &tax.Combo{
				Category: br.TaxCategoryICMS,
				Ext: tax.ExtensionsOf(cbc.CodeMap{
					addon.ExtKeyICMSCST:    "00",
					addon.ExtKeyICMSCSOSN:  "102",
					addon.ExtKeyICMSOrigin: "0",
				}),
			},
			err: "ICMS tax combo cannot include both 'br-nfe-icms-cst' and 'br-nfe-icms-csosn' extensions",
		},
		{
			name: "ICMS no-value CSOSN with non-zero percent",
			tc: &tax.Combo{
				Category: br.TaxCategoryICMS,
				Percent:  num.NewPercentage(1, 2),
				Ext: tax.ExtensionsOf(cbc.CodeMap{
					addon.ExtKeyICMSCSOSN:  "102",
					addon.ExtKeyICMSOrigin: "0",
				}),
			},
			err: "ICMS percent must be zero for CSOSN codes 102, 103, 202, 203, 300, 400 and 500",
		},
		{
			name: "ICMS no-value CSOSN with zero percent",
			tc: &tax.Combo{
				Category: br.TaxCategoryICMS,
				Percent:  &num.PercentageZero,
				Ext: tax.ExtensionsOf(cbc.CodeMap{
					addon.ExtKeyICMSCSOSN:  "500",
					addon.ExtKeyICMSOrigin: "0",
				}),
			},
		},
		{
			name: "ICMS no-value CSOSN with nil percent",
			tc: &tax.Combo{
				Category: br.TaxCategoryICMS,
				Ext: tax.ExtensionsOf(cbc.CodeMap{
					addon.ExtKeyICMSCSOSN:  "102",
					addon.ExtKeyICMSOrigin: "0",
				}),
			},
		},
		{
			name: "ICMS CSOSN 101 with credit rate percent",
			tc: &tax.Combo{
				Category: br.TaxCategoryICMS,
				Percent:  num.NewPercentage(256, 4),
				Ext: tax.ExtensionsOf(cbc.CodeMap{
					addon.ExtKeyICMSCSOSN:  "101",
					addon.ExtKeyICMSOrigin: "0",
				}),
			},
		},
		{
			name: "ICMS no-value CST with non-zero percent",
			tc: &tax.Combo{
				Category: br.TaxCategoryICMS,
				Percent:  num.NewPercentage(5, 2),
				Ext: tax.ExtensionsOf(cbc.CodeMap{
					addon.ExtKeyICMSCST:    "40",
					addon.ExtKeyICMSOrigin: "0",
				}),
			},
			err: "ICMS percent must be zero for CST codes 40, 41, 50 and 60",
		},
		{
			name: "ICMS no-value CST with zero percent",
			tc: &tax.Combo{
				Category: br.TaxCategoryICMS,
				Percent:  &num.PercentageZero,
				Ext: tax.ExtensionsOf(cbc.CodeMap{
					addon.ExtKeyICMSCST:    "60",
					addon.ExtKeyICMSOrigin: "0",
				}),
			},
		},
		{
			name: "ICMS CST 00 with percent",
			tc: &tax.Combo{
				Category: br.TaxCategoryICMS,
				Percent:  num.NewPercentage(18, 2),
				Ext: tax.ExtensionsOf(cbc.CodeMap{
					addon.ExtKeyICMSCST:    "00",
					addon.ExtKeyICMSOrigin: "0",
				}),
			},
		},
		{
			name: "PIS NT CST with non-zero percent",
			tc: &tax.Combo{
				Category: br.TaxCategoryPIS,
				Percent:  num.NewPercentage(165, 4),
				Ext:      tax.ExtensionsOf(cbc.CodeMap{addon.ExtKeyPISCST: "06"}),
			},
			err: "PIS percent must be zero for CST codes 04 to 09",
		},
		{
			name: "PIS NT CST with zero percent",
			tc: &tax.Combo{
				Category: br.TaxCategoryPIS,
				Percent:  &num.PercentageZero,
				Ext:      tax.ExtensionsOf(cbc.CodeMap{addon.ExtKeyPISCST: "06"}),
			},
		},
		{
			name: "PIS CST 49 with non-zero percent",
			tc: &tax.Combo{
				Category: br.TaxCategoryPIS,
				Percent:  num.NewPercentage(165, 4),
				Ext:      tax.ExtensionsOf(cbc.CodeMap{addon.ExtKeyPISCST: "49"}),
			},
		},
		{
			name: "COFINS NT CST with non-zero percent",
			tc: &tax.Combo{
				Category: br.TaxCategoryCOFINS,
				Percent:  num.NewPercentage(760, 4),
				Ext:      tax.ExtensionsOf(cbc.CodeMap{addon.ExtKeyCOFINSCST: "07"}),
			},
			err: "COFINS percent must be zero for CST codes 04 to 09",
		},
		{
			name: "unrelated category is not constrained",
			tc: &tax.Combo{
				Category: br.TaxCategoryIPI,
			},
		},
	}
	for _, ts := range tests {
		t.Run(ts.name, func(t *testing.T) {
			err := rules.Validate(ts.tc, withAddonContext())
			if ts.err == "" {
				assert.NoError(t, err)
			} else if assert.Error(t, err) {
				assert.ErrorContains(t, err, ts.err)
			}
		})
	}
}

func TestTaxComboNormalization(t *testing.T) {
	tests := []struct {
		name   string
		tc     *tax.Combo
		expect map[cbc.Key]cbc.Code // expected codes; "" means the key must be absent
	}{
		{
			name: "ICMS sets default CST and origin",
			tc:   &tax.Combo{Category: br.TaxCategoryICMS},
			expect: map[cbc.Key]cbc.Code{
				addon.ExtKeyICMSCST:    "00",
				addon.ExtKeyICMSCSOSN:  "",
				addon.ExtKeyICMSOrigin: "0",
			},
		},
		{
			name: "ICMS keeps CSOSN and does not add CST",
			tc: &tax.Combo{
				Category: br.TaxCategoryICMS,
				Ext:      tax.ExtensionsOf(cbc.CodeMap{addon.ExtKeyICMSCSOSN: "102"}),
			},
			expect: map[cbc.Key]cbc.Code{
				addon.ExtKeyICMSCSOSN:  "102",
				addon.ExtKeyICMSCST:    "",
				addon.ExtKeyICMSOrigin: "0",
			},
		},
		{
			name: "ICMS does not override CST or origin",
			tc: &tax.Combo{
				Category: br.TaxCategoryICMS,
				Ext: tax.ExtensionsOf(cbc.CodeMap{
					addon.ExtKeyICMSCST:    "40",
					addon.ExtKeyICMSOrigin: "2",
				}),
			},
			expect: map[cbc.Key]cbc.Code{
				addon.ExtKeyICMSCST:    "40",
				addon.ExtKeyICMSOrigin: "2",
			},
		},
		{
			name:   "PIS sets default CST",
			tc:     &tax.Combo{Category: br.TaxCategoryPIS},
			expect: map[cbc.Key]cbc.Code{addon.ExtKeyPISCST: "01"},
		},
		{
			name: "PIS does not override CST",
			tc: &tax.Combo{
				Category: br.TaxCategoryPIS,
				Ext:      tax.ExtensionsOf(cbc.CodeMap{addon.ExtKeyPISCST: "49"}),
			},
			expect: map[cbc.Key]cbc.Code{addon.ExtKeyPISCST: "49"},
		},
		{
			name:   "COFINS sets default CST",
			tc:     &tax.Combo{Category: br.TaxCategoryCOFINS},
			expect: map[cbc.Key]cbc.Code{addon.ExtKeyCOFINSCST: "01"},
		},
		{
			name:   "unrelated category is untouched",
			tc:     &tax.Combo{Category: br.TaxCategoryIPI},
			expect: map[cbc.Key]cbc.Code{addon.ExtKeyICMSCST: ""},
		},
	}
	for _, ts := range tests {
		t.Run(ts.name, func(t *testing.T) {
			norm.Normalize(ts.tc, tax.AddonContext(addon.V4))
			for k, v := range ts.expect {
				assert.Equal(t, v, ts.tc.Ext.Get(k), "ext %s", k)
			}
		})
	}
}

func TestInvoiceRegimeNormalization(t *testing.T) {
	t.Run("defaults supplier regime to normal", func(t *testing.T) {
		inv := &bill.Invoice{
			Addons:   tax.WithAddons(addon.V4),
			Supplier: &org.Party{Name: "Test Supplier"},
		}
		norm.Normalize(inv, tax.AddonContext(addon.V4))
		assert.Equal(t, cbc.Code("3"), inv.Supplier.Ext.Get(addon.ExtKeyRegime))
	})

	t.Run("does not override an existing supplier regime", func(t *testing.T) {
		inv := &bill.Invoice{
			Addons: tax.WithAddons(addon.V4),
			Supplier: &org.Party{
				Name: "Test Supplier",
				Ext:  tax.ExtensionsOf(cbc.CodeMap{addon.ExtKeyRegime: "1"}),
			},
		}
		norm.Normalize(inv, tax.AddonContext(addon.V4))
		assert.Equal(t, cbc.Code("1"), inv.Supplier.Ext.Get(addon.ExtKeyRegime))
	})

	t.Run("nil supplier is a no-op", func(t *testing.T) {
		inv := &bill.Invoice{Addons: tax.WithAddons(addon.V4)}
		assert.NotPanics(t, func() {
			norm.Normalize(inv, tax.AddonContext(addon.V4))
		})
	})
}

// regimeInvoice builds a minimal invoice with the given supplier regime and
// line taxes, suitable for normalization tests.
func regimeInvoice(regime cbc.Code, taxes tax.Set) *bill.Invoice {
	inv := &bill.Invoice{
		Addons:   tax.WithAddons(addon.V4),
		Supplier: &org.Party{Name: "Test Supplier"},
		Lines: []*bill.Line{
			{
				Quantity: num.MakeAmount(1, 0),
				Item: &org.Item{
					Name:  "Test Product",
					Price: num.NewAmount(1000, 2),
				},
				Taxes: taxes,
			},
		},
	}
	if regime != "" {
		inv.Supplier.Ext = tax.ExtensionsOf(cbc.CodeMap{addon.ExtKeyRegime: regime})
	}
	return inv
}

func TestSimplesDefaultCodesRejected(t *testing.T) {
	// The combo code defaults suit the normal regime only: a Simples Nacional
	// supplier omitting the situation codes gets CST 00 defaulted, which the
	// regime cross-validation rejects — codes must be set explicitly under
	// regimes 1 and 4.
	inv := regimeInvoice("1", tax.Set{
		{Category: br.TaxCategoryICMS},
		{Category: br.TaxCategoryPIS},
		{Category: br.TaxCategoryCOFINS},
	})
	norm.Normalize(inv, tax.AddonContext(addon.V4))

	icms := inv.Lines[0].Taxes.Get(br.TaxCategoryICMS)
	assert.Equal(t, cbc.Code("00"), icms.Ext.Get(addon.ExtKeyICMSCST))

	err := rules.Validate(inv)
	assert.ErrorContains(t, err, "Simples Nacional issuers (regime 1 or 4) must use 'br-nfe-icms-csosn'")
}

func TestOrderDeliveryTaxComboNormalization(t *testing.T) {
	// The per-combo defaults apply to every document type carrying tax
	// combos, not only invoices.
	line := func() *bill.Line {
		return &bill.Line{
			Quantity: num.MakeAmount(1, 0),
			Item: &org.Item{
				Name:  "Test Product",
				Price: num.NewAmount(1000, 2),
			},
			Taxes: tax.Set{
				{Category: br.TaxCategoryICMS},
				{Category: br.TaxCategoryPIS},
				{Category: br.TaxCategoryCOFINS},
			},
		}
	}

	t.Run("orders get situation code defaults", func(t *testing.T) {
		ord := &bill.Order{
			Addons:   tax.WithAddons(addon.V4),
			Supplier: &org.Party{Name: "Test Supplier"},
			Lines:    []*bill.Line{line()},
		}
		norm.Normalize(ord, tax.AddonContext(addon.V4))

		icms := ord.Lines[0].Taxes.Get(br.TaxCategoryICMS)
		assert.Equal(t, cbc.Code("00"), icms.Ext.Get(addon.ExtKeyICMSCST))
		assert.Equal(t, cbc.Code("0"), icms.Ext.Get(addon.ExtKeyICMSOrigin))
		pis := ord.Lines[0].Taxes.Get(br.TaxCategoryPIS)
		assert.Equal(t, cbc.Code("01"), pis.Ext.Get(addon.ExtKeyPISCST))
		cofins := ord.Lines[0].Taxes.Get(br.TaxCategoryCOFINS)
		assert.Equal(t, cbc.Code("01"), cofins.Ext.Get(addon.ExtKeyCOFINSCST))
	})

	t.Run("deliveries get situation code defaults", func(t *testing.T) {
		dlv := &bill.Delivery{
			Addons:   tax.WithAddons(addon.V4),
			Supplier: &org.Party{Name: "Test Supplier"},
			Lines:    []*bill.Line{line()},
		}
		norm.Normalize(dlv, tax.AddonContext(addon.V4))

		icms := dlv.Lines[0].Taxes.Get(br.TaxCategoryICMS)
		assert.Equal(t, cbc.Code("00"), icms.Ext.Get(addon.ExtKeyICMSCST))
	})
}
