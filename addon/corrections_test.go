package addon_test

import (
	"testing"

	"github.com/invopop/gobl.br.nfe/addon"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/head"
	"github.com/invopop/gobl/rules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInvoiceCorrection(t *testing.T) {
	t.Run("credit note with sefaz-key stamp", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Code = cbc.Code("NFE-001")

		require.NoError(t, inv.Correct(
			bill.Credit,
			bill.WithStamps(sefazStamps()),
			bill.WithReason("Devolução de mercadoria"),
		))

		assert.Equal(t, bill.InvoiceTypeCreditNote, inv.Type)
		require.Len(t, inv.Preceding, 1)
		pre := inv.Preceding[0]
		require.Len(t, pre.Stamps, 1)
		assert.Equal(t, addon.StampProviderSEFAZKey, pre.Stamps[0].Provider)
		assert.Equal(t, sefazAccessKey, pre.Stamps[0].Value)
		require.NoError(t, rules.Validate(inv))

		// The corrective NF-e purpose is NOT auto-derived from the bill type; it stays
		// normalized to "1" unless the caller sets it explicitly. (Deriving finNFe from
		// the invoice type is an out-of-scope future enhancement.)
		assert.Equal(t, addon.PurposeNormal, inv.Tax.Ext.Get(addon.ExtKeyPurpose))
	})

	t.Run("debit note with sefaz-key stamp", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Code = cbc.Code("NFE-001")

		require.NoError(t, inv.Correct(
			bill.Debit,
			bill.WithStamps(sefazStamps()),
			bill.WithReason("Cobrança complementar"),
		))

		assert.Equal(t, bill.InvoiceTypeDebitNote, inv.Type)
		require.Len(t, inv.Preceding, 1)
		pre := inv.Preceding[0]
		require.Len(t, pre.Stamps, 1)
		assert.Equal(t, addon.StampProviderSEFAZKey, pre.Stamps[0].Provider)
		assert.Equal(t, sefazAccessKey, pre.Stamps[0].Value)
		require.NoError(t, rules.Validate(inv))

		// The corrective NF-e purpose is NOT auto-derived from the bill type; it stays
		// normalized to "1" unless the caller sets it explicitly. (Deriving finNFe from
		// the invoice type is an out-of-scope future enhancement.)
		assert.Equal(t, addon.PurposeNormal, inv.Tax.Ext.Get(addon.ExtKeyPurpose))
	})

	t.Run("credit note without sefaz-key stamp fails", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Code = cbc.Code("NFE-001")

		err := inv.Correct(
			bill.Credit,
			bill.WithReason("Devolução de mercadoria"),
		)
		assert.ErrorContains(t, err, "missing stamp: sefaz-key")
	})

	t.Run("debit note without sefaz-key stamp fails", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Code = cbc.Code("NFE-001")

		err := inv.Correct(
			bill.Debit,
			bill.WithReason("Cobrança complementar"),
		)
		assert.ErrorContains(t, err, "missing stamp: sefaz-key")
	})

	t.Run("credit note without reason succeeds", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Code = cbc.Code("NFE-001")

		require.NoError(t, inv.Correct(
			bill.Credit,
			bill.WithStamps(sefazStamps()),
		))
		require.NoError(t, rules.Validate(inv))
	})
}

const sefazAccessKey = "12345678901234567890123456789012345678901234"

func sefazStamps() []*head.Stamp {
	return []*head.Stamp{
		{
			Provider: addon.StampProviderSEFAZKey,
			Value:    sefazAccessKey,
		},
	}
}
