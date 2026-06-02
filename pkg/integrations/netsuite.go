package integrations

import (
	"errors"
	"net/http"
)

// ErrNotImplemented is returned by skeleton adapters that are registered but not
// yet wired to a live integration.
var ErrNotImplemented = errors.New("integrations: adapter not implemented")

// NetSuite is a skeleton adapter: it reserves the "netsuite" kind and satisfies
// the Adapter contract, but Verify/ParseSales are not yet implemented. It lets
// the marketplace list NetSuite as "coming soon" without special-casing.
type NetSuite struct{}

// NewNetSuite builds a NetSuite skeleton adapter.
func NewNetSuite() NetSuite { return NetSuite{} }

func (NetSuite) Kind() string { return "netsuite" }

// AccountHandle has no header convention yet.
func (NetSuite) AccountHandle(_ []byte, _ *http.Request) string { return "" }

// EventID has no idempotency key yet.
func (NetSuite) EventID(_ []byte, _ *http.Request) string { return "" }

// Verify is not implemented; the skeleton rejects all webhooks for safety.
func (NetSuite) Verify(_ string, _ []byte, _ *http.Request) error { return ErrNotImplemented }

// ParseSales is not implemented.
func (NetSuite) ParseSales(_ []byte) ([]Sale, error) { return nil, ErrNotImplemented }
