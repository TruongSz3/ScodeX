package ipc

import "testing"

func TestValidateLocalAddress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		addr             string
		allowNonLoopback bool
		wantErr          bool
	}{
		{name: "loopback ipv4 allowed", addr: "127.0.0.1:7777", wantErr: false},
		{name: "localhost allowed", addr: "localhost:7777", wantErr: false},
		{name: "loopback ipv6 allowed", addr: "[::1]:7777", wantErr: false},
		{name: "wildcard denied by default", addr: "0.0.0.0:7777", wantErr: true},
		{name: "wildcard allowed with explicit opt-in", addr: "0.0.0.0:7777", allowNonLoopback: true, wantErr: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateLocalAddress(tt.addr, tt.allowNonLoopback)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error for %q", tt.addr)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.addr, err)
			}
		})
	}
}
