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
