package requestscope

import (
	"context"
	"testing"
)

func TestPrincipalRoundTrip(t *testing.T) {
	want := Principal{AccountID: "account-1"}
	got, ok := FromContext(WithPrincipal(context.Background(), want))
	if !ok || got != want {
		t.Fatalf("principal=%+v exists=%v, want %+v", got, ok, want)
	}
}

func TestMissingPrincipalIsExplicit(t *testing.T) {
	if _, ok := FromContext(context.Background()); ok {
		t.Fatal("background context unexpectedly contained a principal")
	}
}

func TestUnrestrictedRequiresMatchingAuthenticatedPrincipal(t *testing.T) {
	for _, test := range []struct {
		name      string
		ctx       context.Context
		accountID string
		want      bool
	}{
		{"missing principal", context.Background(), "admin", false},
		{"ordinary principal", WithPrincipal(context.Background(), Principal{AccountID: "admin"}), "admin", false},
		{"superadmin", WithPrincipal(context.Background(), Principal{AccountID: "admin", Unrestricted: true}), "admin", true},
		{"different account", WithPrincipal(context.Background(), Principal{AccountID: "admin", Unrestricted: true}), "worker", false},
		{"empty identity", WithPrincipal(context.Background(), Principal{Unrestricted: true}), "", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := IsUnrestricted(test.ctx, test.accountID); got != test.want {
				t.Fatalf("IsUnrestricted=%t, want %t", got, test.want)
			}
		})
	}
}
