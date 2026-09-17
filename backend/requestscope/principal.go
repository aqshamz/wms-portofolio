package requestscope

import "context"

type principalKey struct{}

// Principal identifies the authenticated account whose data scope must be
// applied by repositories. Unrestricted is reserved for wildcard
// administrators and deployments where authorization enforcement is disabled.
type Principal struct {
	AccountID    string
	Unrestricted bool
}

func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalKey{}, principal)
}

func FromContext(ctx context.Context) (Principal, bool) {
	principal, ok := ctx.Value(principalKey{}).(Principal)
	return principal, ok
}

// IsUnrestricted only trusts the principal installed by authentication and does
// not permit its privilege to be used when checking a different account.
func IsUnrestricted(ctx context.Context, accountID string) bool {
	principal, exists := FromContext(ctx)
	return exists && principal.Unrestricted && accountID != "" && principal.AccountID == accountID
}
