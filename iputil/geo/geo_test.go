package geo

import (
	"net"
	"testing"
)

func TestAddrFromIP(t *testing.T) {
	tests := []struct {
		name string
		ip   net.IP
		want string
	}{
		{name: "IPv4", ip: net.ParseIP("192.0.2.1"), want: "192.0.2.1"},
		{name: "IPv6", ip: net.ParseIP("2001:db8::1"), want: "2001:db8::1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := addrFromIP(tt.ip)
			if err != nil {
				t.Fatal(err)
			}
			if got.String() != tt.want {
				t.Fatalf("addrFromIP(%q) = %q, want %q", tt.ip, got, tt.want)
			}
		})
	}
}

func TestAddrFromIPRejectsInvalidIP(t *testing.T) {
	if _, err := addrFromIP(nil); err == nil {
		t.Fatal("addrFromIP(nil) returned nil error")
	}
}
