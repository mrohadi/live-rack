package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/live-rack/pkg/domain"
)

// ServiceTokenPrefix marks an opaque service token (vs an OIDC JWT). Service
// tokens are first-class principals with the service role.
const ServiceTokenPrefix = "lrk_"

// HashToken returns the hex SHA-256 of a token; only the hash is stored. Pure.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// IsServiceToken reports whether a bearer value is a service token. Pure.
func IsServiceToken(token string) bool {
	return strings.HasPrefix(token, ServiceTokenPrefix)
}

// ServiceTokenLookup resolves a token hash to its principal, or an error if the
// token is unknown/revoked. Implementations read Postgres.
type ServiceTokenLookup interface {
	ResolveServiceToken(ctx context.Context, hash string) (*domain.Principal, error)
}

// ServiceVerifier authenticates opaque service tokens.
type ServiceVerifier struct {
	lookup ServiceTokenLookup
}

// NewServiceVerifier builds a ServiceVerifier.
func NewServiceVerifier(lookup ServiceTokenLookup) *ServiceVerifier {
	return &ServiceVerifier{lookup: lookup}
}

func bearer(r *http.Request) (string, bool) {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return "", false
	}
	return strings.TrimPrefix(h, "Bearer "), true
}

// VerifyRequest resolves a service-token-bearing request to a service Principal.
func (v *ServiceVerifier) VerifyRequest(r *http.Request) (*domain.Principal, error) {
	tok, ok := bearer(r)
	if !ok || !IsServiceToken(tok) {
		return nil, fmt.Errorf("auth: not a service token")
	}
	p, err := v.lookup.ResolveServiceToken(r.Context(), HashToken(tok))
	if err != nil {
		return nil, fmt.Errorf("auth: service token: %w", err)
	}
	p.Role = domain.RoleService
	return p, nil
}

// CompositeVerifier routes service tokens to the ServiceVerifier and everything
// else to the OIDC verifier.
type CompositeVerifier struct {
	service *ServiceVerifier
	oidc    Verifier
}

// NewCompositeVerifier builds a CompositeVerifier.
func NewCompositeVerifier(service *ServiceVerifier, oidc Verifier) *CompositeVerifier {
	return &CompositeVerifier{service: service, oidc: oidc}
}

// VerifyRequest dispatches by token shape.
func (v *CompositeVerifier) VerifyRequest(r *http.Request) (*domain.Principal, error) {
	if tok, ok := bearer(r); ok && IsServiceToken(tok) {
		return v.service.VerifyRequest(r)
	}
	return v.oidc.VerifyRequest(r)
}
