package addon_test

import (
	"fmt"
	"testing"

	"github.com/invopop/gobl.br.nfe/addon"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/currency"
	"github.com/invopop/gobl/norm"
	"github.com/invopop/gobl/num"
	"github.com/invopop/gobl/org"
	"github.com/invopop/gobl/pay"
	"github.com/invopop/gobl/regimes/br"
	"github.com/invopop/gobl/rules"
	"github.com/invopop/gobl/tax"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInvoicesValidation(t *testing.T) {
	t.Run("validates tax extensions", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Tax = nil
		err := rules.Validate(inv)
		assert.ErrorContains(t, err, "tax is required")

		inv.Tax = &bill.Tax{}
		err = rules.Validate(inv)
		assert.ErrorContains(t, err, "tax requires 'br-nfe-model', 'br-nfe-presence', 'br-nfe-purpose' and 'br-nfe-operation-type' extensions")

		inv.Tax.Ext = tax.ExtensionsOf(cbc.CodeMap{
			addon.ExtKeyModel:         addon.ModelNFe,
			addon.ExtKeyPresence:      addon.PresenceDelivery,
			addon.ExtKeyPurpose:       addon.PurposeNormal,
			addon.ExtKeyOperationType: addon.OperationOutbound,
		})
		err = rules.Validate(inv)
		assert.ErrorContains(t, err, "NF-e invoices do not support '4' for 'br-nfe-presence'")

		inv.Tax.Ext = inv.Tax.Ext.Set(addon.ExtKeyPresence, addon.PresenceInPerson)
		err = rules.Validate(inv)
		assert.NoError(t, err)
	})

	t.Run("validates required notes", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Notes = nil
		err := rules.Validate(inv)
		assert.ErrorContains(t, err, "invoice requires a note with key 'reason' to describe the nature of the operation (natOp)")

		inv.Notes = []*org.Note{nil}
		err = rules.Validate(inv)
		assert.ErrorContains(t, err, "invoice requires a note with key 'reason' to describe the nature of the operation (natOp)")

		inv.Notes[0] = &org.Note{
			Key:  org.NoteKeyGeneral,
			Text: "General note",
		}
		err = rules.Validate(inv)
		assert.ErrorContains(t, err, "invoice requires a note with key 'reason' to describe the nature of the operation (natOp)")

		inv.Notes[0].Key = org.NoteKeyReason
		inv.Notes[0].Text = "1234567890123456789012345678901234567890123456789012345678901" // 61 chars
		err = rules.Validate(inv)
		assert.ErrorContains(t, err, "reason note text must be between 1 and 60 characters")

		inv.Notes[0].Text = "123456789012345678901234567890123456789012345678901234567890" // 60 chars
		err = rules.Validate(inv)
		assert.NoError(t, err)
	})

	t.Run("validates payment when invoice is due", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Payment = nil
		inv.Totals.Due = &num.AmountZero
		err := rules.Validate(inv)
		assert.NoError(t, err)

		inv.Totals.Due = nil
		err = rules.Validate(inv)
		assert.ErrorContains(t, err, "payment is required")

		inv.Totals.Due = num.NewAmount(1, 2)
		err = rules.Validate(inv)
		assert.ErrorContains(t, err, "payment is required")

		inv.Payment = &bill.PaymentDetails{}
		err = rules.Validate(inv)
		assert.ErrorContains(t, err, "payment instructions are required")

		inv.Payment.Instructions = &pay.Instructions{
			Key: pay.MeansKeyCash,
			Ext: tax.ExtensionsOf(cbc.CodeMap{
				addon.ExtKeyPaymentMeans: "01",
			}),
		}
		err = rules.Validate(inv)
		assert.NoError(t, err)
	})

	t.Run("validates invoice totals due field", func(t *testing.T) {
		inv := validCalculatedInvoice(t)

		inv.Totals.Due = num.NewAmount(-1, 2)
		err := rules.Validate(inv)
		assert.ErrorContains(t, err, "due amount must not be negative")

		inv.Totals.Due = &num.AmountZero
		err = rules.Validate(inv)
		assert.NoError(t, err)

		inv.Totals.Due = num.NewAmount(1, 2)
		err = rules.Validate(inv)
		assert.NoError(t, err)
	})

	t.Run("validates NFe presence when model is NFe", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Tax.Ext = inv.Tax.Ext.Set(addon.ExtKeyModel, addon.ModelNFe)
		inv.Tax.Ext = inv.Tax.Ext.Set(addon.ExtKeyPresence, addon.PresenceDelivery)
		err := rules.Validate(inv)
		assert.ErrorContains(t, err, "NF-e invoices do not support '4' for 'br-nfe-presence'")

		inv.Tax.Ext = inv.Tax.Ext.Set(addon.ExtKeyPresence, addon.PresenceInPerson)
		err = rules.Validate(inv)
		assert.NoError(t, err)
	})

	t.Run("validates NFCe presence when model is NFCe", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Customer = nil // For NFCe, customer is optional

		inv.Tax.Ext = inv.Tax.Ext.Set(addon.ExtKeyModel, addon.ModelNFCe)
		inv.Tax.Ext = inv.Tax.Ext.Set(addon.ExtKeyPresence, addon.PresenceNotApplicable)
		err := rules.Validate(inv)
		assert.ErrorContains(t, err, "NFC-e invoices require in-person or delivery for 'br-nfe-presence'")

		inv.Tax.Ext = inv.Tax.Ext.Set(addon.ExtKeyPresence, addon.PresenceInPerson)
		err = rules.Validate(inv)
		assert.NoError(t, err)
	})

	t.Run("validates credit note type when purpose is credit note", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Tax.Ext = inv.Tax.Ext.Set(addon.ExtKeyPurpose, addon.PurposeCreditNote)
		err := rules.Validate(inv)
		assert.ErrorContains(t, err, "credit note invoices require 'br-nfe-credit-note-type' extension")

		inv.Tax.Ext = inv.Tax.Ext.Set(addon.ExtKeyCreditNoteType, "01")
		err = rules.Validate(inv)
		assert.NoError(t, err)
	})

	t.Run("validates debit note type when purpose is debit note", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Tax.Ext = inv.Tax.Ext.Set(addon.ExtKeyPurpose, addon.PurposeDebitNote)
		err := rules.Validate(inv)
		assert.ErrorContains(t, err, "debit note invoices require 'br-nfe-debit-note-type' extension")

		inv.Tax.Ext = inv.Tax.Ext.Set(addon.ExtKeyDebitNoteType, "01")
		err = rules.Validate(inv)
		assert.NoError(t, err)
	})
}

func TestInvoiceSeriesValidation(t *testing.T) {
	tests := []struct {
		series cbc.Code
		err    string
	}{
		{series: "0"},
		{series: "1"},
		{series: "12"},
		{series: "123"},
		{series: "999"},
		{series: "", err: "series is required"},
		{series: "1000", err: "series format is invalid; must be 0 or 1-999"},
		{series: "abc", err: "series format is invalid; must be 0 or 1-999"},
		{series: "012", err: "series format is invalid; must be 0 or 1-999"},
		{series: "00", err: "series format is invalid; must be 0 or 1-999"},
		{series: "-3", err: "series format is invalid; must be 0 or 1-999"},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("validates series %s", tt.series)
		t.Run(name, func(t *testing.T) {
			inv := validCalculatedInvoice(t)
			inv.Series = tt.series
			err := rules.Validate(inv)
			if tt.err != "" {
				assert.ErrorContains(t, err, tt.err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSupplierValidation(t *testing.T) {
	t.Run("nil supplier", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Supplier = nil
		err := rules.Validate(inv)
		// supplier presence is validated at GOBL level, but our rules
		// will still produce errors for nested nil fields - check no panic
		assert.Error(t, err) // GOBL core requires supplier
	})

	t.Run("validates supplier name", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Supplier.Name = ""
		err := rules.Validate(inv)
		assert.ErrorContains(t, err, "supplier name is required")

		inv.Supplier.Name = "Test Company"
		err = rules.Validate(inv)
		assert.NoError(t, err)
	})

	t.Run("validates supplier addresses required", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Supplier.Addresses = nil
		inv.Supplier.Ext = tax.Extensions{} // remove ext to avoid municipality check on nil addresses
		err := rules.Validate(inv)
		assert.ErrorContains(t, err, "supplier must have at least one address")

		inv.Supplier.Addresses = []*org.Address{}
		err = rules.Validate(inv)
		assert.ErrorContains(t, err, "supplier must have at least one address")

		inv.Supplier.Addresses = []*org.Address{nil}
		inv.Supplier.Ext = tax.ExtensionsOf(cbc.CodeMap{
			"br-ibge-municipality": "3304557",
			addon.ExtKeyRegime:     "3",
		})
		err = rules.Validate(inv)
		assert.ErrorContains(t, err, "supplier address must not be empty")

		inv.Supplier.Addresses = []*org.Address{
			{
				Country:  "BR",
				Street:   "Rua Test",
				Number:   "100",
				Locality: "São Paulo",
				State:    "SP",
				Code:     "01310100",
			},
		}
		err = rules.Validate(inv)
		assert.NoError(t, err)
	})

	t.Run("validates supplier tax ID required", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Supplier.TaxID = nil
		err := rules.Validate(inv)
		assert.ErrorContains(t, err, "supplier tax ID is required")

		inv.Supplier.TaxID = &tax.Identity{
			Country: "BR",
		}
		err = rules.Validate(inv)
		assert.ErrorContains(t, err, "supplier tax ID code is required")

		inv.Supplier.TaxID.Code = "55263640000186"
		err = rules.Validate(inv)
		assert.NoError(t, err)
	})

	t.Run("validates supplier municipality extension when addresses exist", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Supplier.Ext = tax.Extensions{}
		err := rules.Validate(inv)
		assert.ErrorContains(t, err, "requires 'br-ibge-municipality' extension when addresses are present")

		inv.Supplier.Ext = tax.ExtensionsOf(cbc.CodeMap{
			"br-ibge-municipality": "3304557",
			addon.ExtKeyRegime:     "3",
		})
		err = rules.Validate(inv)
		assert.NoError(t, err)
	})

	t.Run("validates supplier address fields", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Supplier.Addresses = []*org.Address{
			{
				Street:   "",
				Number:   "100",
				Locality: "São Paulo",
				State:    "SP",
				Code:     "01310100",
			},
		}
		err := rules.Validate(inv)
		assert.ErrorContains(t, err, "supplier address requires a street")

		inv.Supplier.Addresses[0].Street = "Rua Test"
		inv.Supplier.Addresses[0].Number = ""
		err = rules.Validate(inv)
		assert.ErrorContains(t, err, "supplier address requires a number")

		inv.Supplier.Addresses[0].Number = "100"
		inv.Supplier.Addresses[0].Locality = ""
		err = rules.Validate(inv)
		assert.ErrorContains(t, err, "supplier address requires a locality")

		inv.Supplier.Addresses[0].Locality = "São Paulo"
		inv.Supplier.Addresses[0].State = ""
		err = rules.Validate(inv)
		assert.ErrorContains(t, err, "supplier address requires a state")

		inv.Supplier.Addresses[0].State = "SP"
		inv.Supplier.Addresses[0].Code = ""
		err = rules.Validate(inv)
		assert.ErrorContains(t, err, "supplier address requires a postal code")
	})
}

func TestCustomerValidation(t *testing.T) {
	t.Run("validates customer required for NFe", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Tax.Ext = inv.Tax.Ext.Set(addon.ExtKeyModel, addon.ModelNFe)
		inv.Customer = nil
		err := rules.Validate(inv)
		assert.ErrorContains(t, err, "customer is required for NF-e")

		inv.Customer = &org.Party{
			Name: "Test Customer",
			TaxID: &tax.Identity{
				Country: "BR",
				Code:    "05700736000196",
			},
			Addresses: []*org.Address{
				{
					Country:  "BR",
					Street:   "Rua das Flores",
					Number:   "123",
					Locality: "São Paulo",
					State:    "SP",
					Code:     "01310000",
				},
			},
			Ext: tax.ExtensionsOf(cbc.CodeMap{
				"br-ibge-municipality":  "3550308",
				addon.ExtKeyStateRegInd: addon.StateRegIndNonTaxpayer,
			}),
		}
		err = rules.Validate(inv)
		assert.NoError(t, err)
	})

	t.Run("validates customer addresses required for NFe", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Tax.Ext = inv.Tax.Ext.Set(addon.ExtKeyModel, addon.ModelNFe)
		inv.Customer.Addresses = nil
		inv.Customer.Ext = tax.Extensions{} // avoid municipality check
		err := rules.Validate(inv)
		assert.ErrorContains(t, err, "customer must have at least one address for NF-e")

		inv.Customer.Addresses = []*org.Address{}
		err = rules.Validate(inv)
		assert.ErrorContains(t, err, "customer must have at least one address for NF-e")

		inv.Customer.Addresses = []*org.Address{nil}
		inv.Customer.Ext = tax.ExtensionsOf(cbc.CodeMap{
			"br-ibge-municipality":  "3550308",
			addon.ExtKeyStateRegInd: addon.StateRegIndNonTaxpayer,
		})
		err = rules.Validate(inv)
		assert.ErrorContains(t, err, "customer address must not be empty")

		inv.Customer.Addresses = []*org.Address{
			{
				Country:  "BR",
				Street:   "Rua das Flores",
				Number:   "123",
				Locality: "São Paulo",
				State:    "SP",
				Code:     "01310000",
			},
		}
		err = rules.Validate(inv)
		assert.NoError(t, err)
	})

	t.Run("customer not required for NFCe", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Tax.Ext = inv.Tax.Ext.Set(addon.ExtKeyModel, addon.ModelNFCe)
		inv.Tax.Ext = inv.Tax.Ext.Set(addon.ExtKeyPresence, addon.PresenceInPerson)
		inv.Customer = nil
		err := rules.Validate(inv)
		assert.NoError(t, err)
	})

	t.Run("validates customer tax ID required", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Customer.TaxID = nil
		err := rules.Validate(inv)
		assert.ErrorContains(t, err, "invoice customer must have a tax ID or a foreign country identity")

		// A valid foreign identity allows TaxID to be nil
		inv.Customer.Identities = []*org.Identity{
			{
				Key:     org.IdentityKeyForeign,
				Country: "US",
				Code:    "US-FOREIGN-123",
			},
		}
		err = rules.Validate(inv)
		assert.NoError(t, err)
		inv.Customer.Identities = nil

		// TaxID present but no code still fails rule 15
		inv.Customer.TaxID = &tax.Identity{
			Country: "BR",
		}
		err = rules.Validate(inv)
		assert.ErrorContains(t, err, "customer tax ID code is required")

		inv.Customer.TaxID.Code = "05700736000196"
		err = rules.Validate(inv)
		assert.NoError(t, err)
	})

	t.Run("validates customer municipality when addresses exist", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Customer.Ext = tax.ExtensionsOf(cbc.CodeMap{
			addon.ExtKeyStateRegInd: addon.StateRegIndNonTaxpayer,
		})
		err := rules.Validate(inv)
		assert.ErrorContains(t, err, "requires 'br-ibge-municipality' extension when addresses are present")

		inv.Customer.Ext = tax.ExtensionsOf(cbc.CodeMap{
			"br-ibge-municipality":  "3550308",
			addon.ExtKeyStateRegInd: addon.StateRegIndNonTaxpayer,
		})
		err = rules.Validate(inv)
		assert.NoError(t, err)
	})

	t.Run("validates customer address fields", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Customer.Addresses = []*org.Address{
			{
				Street:   "",
				Number:   "123",
				Locality: "São Paulo",
				State:    "SP",
				Code:     "01310000",
			},
		}
		err := rules.Validate(inv)
		assert.ErrorContains(t, err, "customer address requires a street")

		inv.Customer.Addresses[0].Street = "Rua das Flores"
		inv.Customer.Addresses[0].Number = ""
		err = rules.Validate(inv)
		assert.ErrorContains(t, err, "customer address requires a number")

		inv.Customer.Addresses[0].Number = "123"
		inv.Customer.Addresses[0].Locality = ""
		err = rules.Validate(inv)
		assert.ErrorContains(t, err, "customer address requires a locality")

		inv.Customer.Addresses[0].Locality = "São Paulo"
		inv.Customer.Addresses[0].State = ""
		err = rules.Validate(inv)
		assert.ErrorContains(t, err, "customer address requires a state")

		inv.Customer.Addresses[0].State = "SP"
		inv.Customer.Addresses[0].Code = ""
		err = rules.Validate(inv)
		assert.ErrorContains(t, err, "customer address requires a postal code")
	})

	t.Run("NF-e lines require CFOP extension", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Lines[0].Ext = tax.Extensions{}
		err := rules.Validate(inv)
		assert.ErrorContains(t, err, fmt.Sprintf("invoice lines require '%s' extension", addon.ExtKeyCFOP))

		inv.Lines[0].Ext = tax.ExtensionsOf(cbc.CodeMap{
			addon.ExtKeyCFOP: "5102",
		})
		err = rules.Validate(inv)
		assert.NoError(t, err)
	})

	t.Run("NFC-e lines require CFOP extension", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Tax.Ext = inv.Tax.Ext.Set(addon.ExtKeyModel, addon.ModelNFCe)
		inv.Tax.Ext = inv.Tax.Ext.Set(addon.ExtKeyPresence, addon.PresenceInPerson)
		inv.Customer = nil
		inv.Lines[0].Ext = tax.Extensions{}
		err := rules.Validate(inv)
		assert.ErrorContains(t, err, fmt.Sprintf("invoice lines require '%s' extension", addon.ExtKeyCFOP))

		inv.Lines[0].Ext = tax.ExtensionsOf(cbc.CodeMap{
			addon.ExtKeyCFOP: "5102",
		})
		err = rules.Validate(inv)
		assert.NoError(t, err)
	})
}

func TestCustomerStateRegIndValidation(t *testing.T) {
	stateReg := &org.Identity{
		Key:  addon.IdentityKeyStateReg,
		Code: "112233445566",
	}

	t.Run("requires the extension when a customer is present", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Customer.Ext = inv.Customer.Ext.Delete(addon.ExtKeyStateRegInd)
		err := rules.Validate(inv)
		assert.ErrorContains(t, err, "invoice customer requires 'br-nfe-state-reg-ind' extension")

		inv.Customer.Ext = inv.Customer.Ext.Set(addon.ExtKeyStateRegInd, addon.StateRegIndNonTaxpayer)
		err = rules.Validate(inv)
		assert.NoError(t, err)
	})

	t.Run("does not require the extension without a customer", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Tax.Ext = inv.Tax.Ext.Set(addon.ExtKeyModel, addon.ModelNFCe)
		inv.Customer = nil
		err := rules.Validate(inv)
		assert.NoError(t, err)
	})

	t.Run("validates the extension code", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Customer.Ext = inv.Customer.Ext.Set(addon.ExtKeyStateRegInd, "3")
		err := rules.Validate(inv)
		assert.ErrorContains(t, err, "'br-nfe-state-reg-ind' extension, when set, must be a valid code")
	})

	t.Run("taxpayer requires the state registration", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Customer.Identities = nil
		inv.Customer.Ext = inv.Customer.Ext.Set(addon.ExtKeyStateRegInd, addon.StateRegIndTaxpayer)
		err := rules.Validate(inv)
		assert.ErrorContains(t, err, "invoice customer requires 'br-nfe-state-reg' identity when 'br-nfe-state-reg-ind' is '1'")

		inv.Customer.Identities = []*org.Identity{stateReg}
		err = rules.Validate(inv)
		assert.NoError(t, err)
	})

	t.Run("exempt taxpayer must not have a state registration", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Customer.Identities = []*org.Identity{stateReg}
		inv.Customer.Ext = inv.Customer.Ext.Set(addon.ExtKeyStateRegInd, addon.StateRegIndExempt)
		err := rules.Validate(inv)
		assert.ErrorContains(t, err, "invoice customer must not have 'br-nfe-state-reg' identity when 'br-nfe-state-reg-ind' is '2'")

		inv.Customer.Identities = nil
		err = rules.Validate(inv)
		assert.NoError(t, err)
	})

	t.Run("non-taxpayer may or may not have a state registration", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Customer.Ext = inv.Customer.Ext.Set(addon.ExtKeyStateRegInd, addon.StateRegIndNonTaxpayer)
		inv.Customer.Identities = nil
		err := rules.Validate(inv)
		assert.NoError(t, err)

		inv.Customer.Identities = []*org.Identity{stateReg}
		err = rules.Validate(inv)
		assert.NoError(t, err)
	})

	t.Run("NFC-e customers must be non-taxpayers", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Tax.Ext = inv.Tax.Ext.Set(addon.ExtKeyModel, addon.ModelNFCe)
		inv.Customer.Identities = []*org.Identity{stateReg}
		inv.Customer.Ext = inv.Customer.Ext.Set(addon.ExtKeyStateRegInd, addon.StateRegIndTaxpayer)
		err := rules.Validate(inv)
		assert.ErrorContains(t, err, "NFC-e invoices require '9' for 'br-nfe-state-reg-ind'")

		inv.Customer.Ext = inv.Customer.Ext.Set(addon.ExtKeyStateRegInd, addon.StateRegIndExempt)
		inv.Customer.Identities = nil
		err = rules.Validate(inv)
		assert.ErrorContains(t, err, "NFC-e invoices require '9' for 'br-nfe-state-reg-ind'")

		// The state registration is irrelevant for non-taxpayers, also in NFC-e
		inv.Customer.Ext = inv.Customer.Ext.Set(addon.ExtKeyStateRegInd, addon.StateRegIndNonTaxpayer)
		inv.Customer.Identities = []*org.Identity{stateReg}
		err = rules.Validate(inv)
		assert.NoError(t, err)
	})
}

func TestInvoiceModelNormalization(t *testing.T) {
	t.Run("defaults to NF-e", func(t *testing.T) {
		inv := &bill.Invoice{Addons: tax.WithAddons(addon.V4)}
		norm.Normalize(inv, tax.AddonContext(addon.V4))
		assert.Equal(t, addon.ModelNFe, inv.Tax.GetExt(addon.ExtKeyModel))
	})

	t.Run("sets NFC-e for simplified invoices", func(t *testing.T) {
		inv := &bill.Invoice{Addons: tax.WithAddons(addon.V4)}
		inv.SetTags(tax.TagSimplified)
		norm.Normalize(inv, tax.AddonContext(addon.V4))
		assert.Equal(t, addon.ModelNFCe, inv.Tax.GetExt(addon.ExtKeyModel))
	})

	t.Run("overrides an explicit model", func(t *testing.T) {
		inv := &bill.Invoice{
			Addons: tax.WithAddons(addon.V4),
			Type:   bill.InvoiceTypeStandard,
			Tax:    &bill.Tax{Ext: tax.ExtensionsOf(cbc.CodeMap{addon.ExtKeyModel: addon.ModelNFCe})},
		}
		norm.Normalize(inv, tax.AddonContext(addon.V4))
		assert.Equal(t, addon.ModelNFe, inv.Tax.GetExt(addon.ExtKeyModel))
	})

	t.Run("sets the model before the customer defaults", func(t *testing.T) {
		inv := &bill.Invoice{
			Addons: tax.WithAddons(addon.V4),
			Type:   bill.InvoiceTypeStandard,
			Customer: &org.Party{
				Identities: []*org.Identity{{Key: addon.IdentityKeyStateReg, Code: "112233445566"}},
			},
		}
		inv.SetTags(tax.TagSimplified)
		norm.Normalize(inv, tax.AddonContext(addon.V4))
		assert.Equal(t, addon.ModelNFCe, inv.Tax.GetExt(addon.ExtKeyModel))
		assert.Equal(t, addon.StateRegIndNonTaxpayer, inv.Customer.Ext.Get(addon.ExtKeyStateRegInd))
	})
}

func TestCustomerStateRegIndNormalization(t *testing.T) {
	stateReg := &org.Identity{
		Key:  addon.IdentityKeyStateReg,
		Code: "112233445566",
	}

	t.Run("defaults to taxpayer when the customer has a state registration", func(t *testing.T) {
		inv := validInvoice()
		inv.Customer.Identities = []*org.Identity{stateReg}
		require.NoError(t, inv.Calculate())
		assert.Equal(t, addon.StateRegIndTaxpayer, inv.Customer.Ext.Get(addon.ExtKeyStateRegInd))
	})

	t.Run("defaults to non-taxpayer when the customer has no state registration", func(t *testing.T) {
		inv := validInvoice()
		inv.Customer.Identities = nil
		require.NoError(t, inv.Calculate())
		assert.Equal(t, addon.StateRegIndNonTaxpayer, inv.Customer.Ext.Get(addon.ExtKeyStateRegInd))
	})

	t.Run("defaults to non-taxpayer for NFC-e regardless of the state registration", func(t *testing.T) {
		inv := validInvoice()
		inv.Tax = nil // model to be set during normalization
		inv.SetTags(tax.TagSimplified)
		inv.Customer.Identities = []*org.Identity{stateReg}
		require.NoError(t, inv.Calculate())
		assert.Equal(t, addon.ModelNFCe, inv.Tax.Ext.Get(addon.ExtKeyModel))
		assert.Equal(t, addon.StateRegIndNonTaxpayer, inv.Customer.Ext.Get(addon.ExtKeyStateRegInd))
	})

	t.Run("does not override an existing value", func(t *testing.T) {
		inv := validInvoice()
		inv.Customer.Identities = nil
		inv.Customer.Ext = inv.Customer.Ext.Set(addon.ExtKeyStateRegInd, addon.StateRegIndExempt)
		require.NoError(t, inv.Calculate())
		assert.Equal(t, addon.StateRegIndExempt, inv.Customer.Ext.Get(addon.ExtKeyStateRegInd))
	})

	t.Run("nil customer is a no-op", func(t *testing.T) {
		inv := validInvoice()
		inv.Customer = nil
		assert.NotPanics(t, func() {
			require.NoError(t, inv.Calculate())
		})
		assert.Nil(t, inv.Customer)
	})
}

func TestAddressCountryValidation(t *testing.T) {
	t.Run("supplier address requires a country", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Supplier.Addresses[0].Country = ""
		err := rules.Validate(inv)
		assert.ErrorContains(t, err, "supplier address requires a country")

		inv.Supplier.Addresses[0].Country = "BR"
		err = rules.Validate(inv)
		assert.NoError(t, err)
	})

	t.Run("customer address requires a country", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Customer.Addresses[0].Country = ""
		err := rules.Validate(inv)
		assert.ErrorContains(t, err, "customer address requires a country")

		inv.Customer.Addresses[0].Country = "BR"
		err = rules.Validate(inv)
		assert.NoError(t, err)
	})
}

func TestInvoiceCurrencyValidation(t *testing.T) {
	t.Run("non-BRL currency without exchange rates", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Currency = "USD"
		require.NoError(t, inv.Calculate())
		err := rules.Validate(inv)
		assert.ErrorContains(t, err, "[GOBL-BR-NFE-BILL-INVOICE-34] invoice currency must be BRL or provide exchange rate for conversion")
	})

	t.Run("non-BRL currency with exchange rates", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Currency = "USD"
		inv.ExchangeRates = []*currency.ExchangeRate{
			{
				From:   "USD",
				To:     "BRL",
				Amount: num.MakeAmount(500, 2),
			},
		}
		require.NoError(t, inv.Calculate())
		err := rules.Validate(inv)
		assert.NoError(t, err)
	})
}

func TestICMSRegimeCodeValidation(t *testing.T) {
	// Sets a line's ICMS combo extensions and percent in place.
	setICMS := func(inv *bill.Invoice, ext cbc.CodeMap, percent *num.Percentage) {
		icms := inv.Lines[0].Taxes.Get(br.TaxCategoryICMS)
		icms.Ext = tax.ExtensionsOf(ext)
		icms.Percent = percent
	}

	t.Run("Simples supplier cannot use a non-fuel CST", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Supplier.Ext = inv.Supplier.Ext.Set(addon.ExtKeyRegime, "1")
		err := rules.Validate(inv)
		assert.ErrorContains(t, err, "Simples Nacional issuers (regime 1 or 4) must use 'br-nfe-icms-csosn'; 'br-nfe-icms-cst' is only allowed for fuel codes 02, 15, 53 and 61")
	})

	t.Run("MEI supplier cannot use a non-fuel CST", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Supplier.Ext = inv.Supplier.Ext.Set(addon.ExtKeyRegime, "4")
		err := rules.Validate(inv)
		assert.ErrorContains(t, err, "Simples Nacional issuers (regime 1 or 4) must use 'br-nfe-icms-csosn'")
	})

	t.Run("Simples supplier may use a fuel CST", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Supplier.Ext = inv.Supplier.Ext.Set(addon.ExtKeyRegime, "1")
		setICMS(inv, cbc.CodeMap{
			addon.ExtKeyICMSCST:    "02",
			addon.ExtKeyICMSOrigin: "0",
		}, num.NewPercentage(18, 2))
		err := rules.Validate(inv)
		assert.NoError(t, err)
	})

	t.Run("Simples supplier with CSOSN is valid", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Supplier.Ext = inv.Supplier.Ext.Set(addon.ExtKeyRegime, "1")
		setICMS(inv, cbc.CodeMap{
			addon.ExtKeyICMSCSOSN:  "102",
			addon.ExtKeyICMSOrigin: "0",
		}, &num.PercentageZero)
		err := rules.Validate(inv)
		assert.NoError(t, err)
	})

	t.Run("normal regime supplier cannot use CSOSN", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		setICMS(inv, cbc.CodeMap{
			addon.ExtKeyICMSCSOSN:  "102",
			addon.ExtKeyICMSOrigin: "0",
		}, &num.PercentageZero)
		err := rules.Validate(inv)
		assert.ErrorContains(t, err, "only Simples Nacional issuers can use 'br-nfe-icms-csosn'; others must use 'br-nfe-icms-cst'")
	})

	t.Run("regime 2 supplier cannot use CSOSN", func(t *testing.T) {
		inv := validCalculatedInvoice(t)
		inv.Supplier.Ext = inv.Supplier.Ext.Set(addon.ExtKeyRegime, "2")
		setICMS(inv, cbc.CodeMap{
			addon.ExtKeyICMSCSOSN:  "102",
			addon.ExtKeyICMSOrigin: "0",
		}, &num.PercentageZero)
		err := rules.Validate(inv)
		assert.ErrorContains(t, err, "only Simples Nacional issuers can use 'br-nfe-icms-csosn'; others must use 'br-nfe-icms-cst'")
	})

	t.Run("ignores combos in discounts and charges", func(t *testing.T) {
		// The converter never maps document-level discount or charge taxes,
		// so their situation codes are not regime-checked.
		inv := validCalculatedInvoice(t)
		inv.Supplier.Ext = inv.Supplier.Ext.Set(addon.ExtKeyRegime, "1")
		setICMS(inv, cbc.CodeMap{
			addon.ExtKeyICMSCSOSN:  "102",
			addon.ExtKeyICMSOrigin: "0",
		}, &num.PercentageZero)
		inv.Discounts = []*bill.Discount{
			{
				Reason: "Promotional discount",
				Amount: num.MakeAmount(100, 2),
				Taxes: tax.Set{
					{
						Category: br.TaxCategoryICMS,
						Percent:  num.NewPercentage(18, 2),
						Ext: tax.ExtensionsOf(cbc.CodeMap{
							addon.ExtKeyICMSCST:    "00",
							addon.ExtKeyICMSOrigin: "0",
						}),
					},
				},
			},
		}
		assert.NoError(t, rules.Validate(inv))
	})
}

// validInvoice creates a raw invoice suitable for scenario tests that call Calculate() themselves.
func validInvoice() *bill.Invoice {
	return &bill.Invoice{
		Addons:   tax.WithAddons(addon.V4),
		Currency: "BRL",
		Series:   cbc.Code("123"),
		Supplier: &org.Party{
			Name: "Test Supplier LTDA",
			TaxID: &tax.Identity{
				Country: "BR",
				Code:    "55263640000186",
			},
			Identities: []*org.Identity{
				{
					Key:  addon.IdentityKeyStateReg,
					Code: "35503304557308",
				},
			},
			Addresses: []*org.Address{
				{
					Street:   "Av Paulista",
					Number:   "1578",
					Locality: "São Paulo",
					State:    "SP",
					Code:     "01310100",
				},
			},
			Ext: tax.ExtensionsOf(cbc.CodeMap{
				"br-ibge-municipality": "3304557",
			}),
		},
		Customer: &org.Party{
			Name: "Test Customer",
			TaxID: &tax.Identity{
				Country: "BR",
				Code:    "05700736000196",
			},
			Addresses: []*org.Address{
				{
					Country:  "BR",
					Street:   "Rua das Flores",
					Number:   "123",
					Locality: "São Paulo",
					State:    "SP",
					Code:     "01310000",
				},
			},
			Ext: tax.ExtensionsOf(cbc.CodeMap{
				"br-ibge-municipality": "3550308",
			}),
		},
		Tax: &bill.Tax{
			Ext: tax.ExtensionsOf(cbc.CodeMap{
				addon.ExtKeyModel:    addon.ModelNFe,
				addon.ExtKeyPresence: addon.PresenceInPerson,
			}),
		},
		Notes: []*org.Note{
			{
				Key:  org.NoteKeyReason,
				Text: "VENDA DE MERCADORIA",
			},
		},
		Lines: []*bill.Line{
			{
				Quantity: num.MakeAmount(1, 0),
				Item: &org.Item{
					Name:  "Test Product",
					Price: num.NewAmount(10000, 2),
				},
				Taxes: tax.Set{
					{
						Category: br.TaxCategoryICMS,
						Percent:  num.NewPercentage(18, 2),
					},
					{
						Category: br.TaxCategoryPIS,
						Percent:  num.NewPercentage(165, 4),
					},
					{
						Category: br.TaxCategoryCOFINS,
						Percent:  num.NewPercentage(760, 4),
					},
				},
				Ext: tax.ExtensionsOf(cbc.CodeMap{
					addon.ExtKeyCFOP: "5102",
				}),
			},
		},
		Payment: &bill.PaymentDetails{
			Instructions: &pay.Instructions{
				Key: pay.MeansKeyCash,
			},
		},
	}
}

// validCalculatedInvoice creates a fully valid invoice with Calculate() applied,
// suitable for post-modification testing of specific rule violations.
func validCalculatedInvoice(t *testing.T) *bill.Invoice {
	t.Helper()
	inv := &bill.Invoice{
		Addons:   tax.WithAddons(addon.V4),
		Currency: "BRL",
		Series:   cbc.Code("123"),
		Supplier: &org.Party{
			Name: "Test Supplier LTDA",
			TaxID: &tax.Identity{
				Country: "BR",
				Code:    "55263640000186",
			},
			Identities: []*org.Identity{
				{
					Key:  addon.IdentityKeyStateReg,
					Code: "35503304557308",
				},
			},
			Addresses: []*org.Address{
				{
					Street:   "Av Paulista",
					Number:   "1578",
					Locality: "São Paulo",
					State:    "SP",
					Code:     "01310100",
				},
			},
			Ext: tax.ExtensionsOf(cbc.CodeMap{
				"br-ibge-municipality": "3304557",
			}),
		},
		Customer: &org.Party{
			Name: "Test Customer",
			TaxID: &tax.Identity{
				Country: "BR",
				Code:    "05700736000196",
			},
			Addresses: []*org.Address{
				{
					Country:  "BR",
					Street:   "Rua das Flores",
					Number:   "123",
					Locality: "São Paulo",
					State:    "SP",
					Code:     "01310000",
				},
			},
			Ext: tax.ExtensionsOf(cbc.CodeMap{
				"br-ibge-municipality": "3550308",
			}),
		},
		Notes: []*org.Note{
			{
				Key:  org.NoteKeyReason,
				Text: "VENDA DE MERCADORIA",
			},
		},
		Lines: []*bill.Line{
			{
				Quantity: num.MakeAmount(1, 0),
				Item: &org.Item{
					Name:  "Test Product",
					Price: num.NewAmount(10000, 2),
				},
				Taxes: tax.Set{
					{
						Category: br.TaxCategoryICMS,
						Percent:  num.NewPercentage(18, 2),
					},
					{
						Category: br.TaxCategoryPIS,
						Percent:  num.NewPercentage(165, 4),
					},
					{
						Category: br.TaxCategoryCOFINS,
						Percent:  num.NewPercentage(760, 4),
					},
				},
				Ext: tax.ExtensionsOf(cbc.CodeMap{
					addon.ExtKeyCFOP: "5102",
				}),
			},
		},
		Payment: &bill.PaymentDetails{
			Instructions: &pay.Instructions{
				Key: pay.MeansKeyCash,
			},
		},
	}
	require.NoError(t, inv.Calculate())
	// Presence is not set by scenarios, set it manually after Calculate
	inv.Tax.Ext = inv.Tax.Ext.Set(addon.ExtKeyPresence, addon.PresenceInPerson)
	require.NoError(t, rules.Validate(inv))
	return inv
}
