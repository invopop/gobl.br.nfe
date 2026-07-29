package addon_test

import (
	"testing"

	"github.com/invopop/gobl.br.nfe/addon"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/org"
	"github.com/invopop/gobl/rules"
	"github.com/stretchr/testify/assert"
)

func TestCESTIdentityValidation(t *testing.T) {
	tests := []struct {
		name string
		code cbc.Code
		err  string
	}{
		{name: "valid code", code: "0300700"},
		{name: "too short", code: "123", err: "CEST identity code must be a 7-digit number"},
		{name: "too long", code: "12345678", err: "CEST identity code must be a 7-digit number"},
		{name: "non-numeric", code: "03A0700", err: "CEST identity code must be a 7-digit number"},
	}
	for _, ts := range tests {
		t.Run(ts.name, func(t *testing.T) {
			id := &org.Identity{
				Key:  addon.IdentityKeyCEST,
				Code: ts.code,
			}
			err := rules.Validate(id, withAddonContext())
			if ts.err == "" {
				assert.NoError(t, err)
			} else if assert.Error(t, err) {
				assert.ErrorContains(t, err, ts.err)
			}
		})
	}

	t.Run("other identities are not constrained", func(t *testing.T) {
		id := &org.Identity{
			Key:  addon.IdentityKeyStateReg,
			Code: "35503304557308",
		}
		assert.NoError(t, rules.Validate(id, withAddonContext()))
	})
}

func TestNCMIdentityValidation(t *testing.T) {
	tests := []struct {
		name string
		code cbc.Code
		err  string
	}{
		{name: "valid 8-digit code", code: "94036000"},
		{name: "valid 2-digit chapter", code: "94"},
		{name: "dotted code", code: "9403.60.00", err: "NCM identity code must be an 8-digit number"},
		{name: "wrong length", code: "9403600", err: "NCM identity code must be an 8-digit number"},
	}
	for _, ts := range tests {
		t.Run(ts.name, func(t *testing.T) {
			id := &org.Identity{
				Key:  org.IdentityKeyNCM,
				Code: ts.code,
			}
			err := rules.Validate(id, withAddonContext())
			if ts.err == "" {
				assert.NoError(t, err)
			} else if assert.Error(t, err) {
				assert.ErrorContains(t, err, ts.err)
			}
		})
	}
}
