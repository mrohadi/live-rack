package auth

import (
	"context"
	"fmt"

	"github.com/live-rack/pkg/domain"
)

type contextKey string

const principalKey contextKey = "principal"

func WithPrincipal(ctx context.Context, p *domain.Principal) context.Context {
	return context.WithValue(ctx, principalKey, p)
}

func PrincipalFrom(ctx context.Context) (*domain.Principal, error) {
	p, ok := ctx.Value(principalKey).(*domain.Principal)
	if !ok || p == nil {
		return nil, fmt.Errorf("auth: no principal in context")
	}
	return p, nil
}

func MustPrincipal(ctx context.Context) *domain.Principal {
	p, err := PrincipalFrom(ctx)
	if err != nil {
		panic(err)
	}
	return p
}
