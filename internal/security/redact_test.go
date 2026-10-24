package security

import (
	"strings"
	"testing"
)

func TestRedactLine(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "password",
			in:   " password cisco",
			want: " password <redacted>",
		},
		{
			name: "enable secret",
			in:   "enable secret 5 $1$abc",
			want: "enable secret <redacted>",
		},
		{
			name: "username password",
			in:   "username admin privilege 15 password 7 0822455D0A16",
			want: "username admin privilege 15 password <redacted>",
		},
		{
			name: "username secret",
			in:   " username admin secret 5 $1$abc",
			want: " username admin secret <redacted>",
		},
		{
			name: "snmp community",
			in:   "snmp-server community public RO",
			want: "snmp-server community <redacted>",
		},
		{
			name: "crypto key",
			in:   "crypto key generate rsa modulus 2048",
			want: "crypto key <redacted>",
		},
		{
			name: "key-string",
			in:   " key-string 7 1234567890",
			want: " key-string <redacted>",
		},
		{
			name: "pre-shared-key",
			in:   "pre-shared-key address 1.2.3.4 key 0 shared",
			want: "pre-shared-key <redacted>",
		},
		{
			name: "wpa-psk",
			in:   "wpa-psk ascii 0 wireless-secret",
			want: "wpa-psk <redacted>",
		},
		{
			name: "snmp host community",
			in:   "snmp-server host 10.0.0.1 version 2c public",
			want: "snmp-server host 10.0.0.1 version 2c <redacted>",
		},
		{
			name: "snmp host community without version",
			in:   "snmp-server host 10.0.0.1 private",
			want: "snmp-server host 10.0.0.1 <redacted>",
		},
		{
			name: "snmp host informs community",
			in:   "snmp-server host 10.0.0.1 informs version 2c public",
			want: "snmp-server host 10.0.0.1 informs version 2c <redacted>",
		},
		{
			name: "snmp v3 auth and priv",
			in:   "snmp-server user admin netmon v3 auth sha authpass priv aes 128 privpass",
			want: "snmp-server user admin netmon v3 auth sha <redacted>",
		},
		{
			name: "snmp v3 priv only",
			in:   "snmp-server user admin netmon v3 priv aes 128 privpass",
			want: "snmp-server user admin netmon v3 priv aes 128 <redacted>",
		},
		{
			name: "snmp v3 sha-384",
			in:   "snmp-server user a1 g v3 auth sha-384 leak384",
			want: "snmp-server user a1 g v3 auth sha-384 <redacted>",
		},
		{
			name: "snmp v3 sha-224",
			in:   "snmp-server user a2 g v3 auth sha-224 leak224",
			want: "snmp-server user a2 g v3 auth sha-224 <redacted>",
		},
		{
			name: "snmp v3 des56",
			in:   "snmp-server user a3 g v3 priv des56 leakDes56",
			want: "snmp-server user a3 g v3 priv des56 <redacted>",
		},
		{
			name: "case insensitive",
			in:   "SNMP-SERVER COMMUNITY Public RO",
			want: "SNMP-SERVER COMMUNITY <redacted>",
		},
		{
			name: "password-encryption is not a secret",
			in:   "no service password-encryption",
			want: "no service password-encryption",
		},
		{
			name: "service password-encryption is not a secret",
			in:   "service password-encryption",
			want: "service password-encryption",
		},
		{
			name: "plain line",
			in:   "interface FastEthernet0/1",
			want: "interface FastEthernet0/1",
		},
		{
			name: "empty line",
			in:   "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RedactLine(tt.in); got != tt.want {
				t.Fatalf("RedactLine(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestRedactLineNeverLeavesTheSecretValue(t *testing.T) {
	lines := []string{
		"enable secret 5 $1$super-secret",
		"username admin password cisco",
		"snmp-server community private RO",
		"crypto key generate rsa",
	}
	for _, line := range lines {
		got := RedactLine(line)
		if !strings.Contains(got, redacted) {
			t.Fatalf("RedactLine(%q) = %q, missing redaction marker", line, got)
		}
		for _, secret := range []string{"$1$super-secret", "cisco", "private"} {
			if strings.Contains(got, secret) {
				t.Fatalf("RedactLine(%q) leaked %q: %q", line, secret, got)
			}
		}
	}
}

func TestRedactLines(t *testing.T) {
	in := []string{
		"!",
		" enable secret 5 hash",
		"interface Fa0/1",
		" snmp-server community public RO",
	}
	want := []string{
		"!",
		" enable secret <redacted>",
		"interface Fa0/1",
		" snmp-server community <redacted>",
	}
	got := RedactLines(in)
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("RedactLines = %q, want %q", got, want)
	}
}
