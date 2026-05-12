package httpapi

import (
	"reflect"
	"testing"
)

// TestTrustedProxiesFromEnv covers the DEVHUB_TRUSTED_PROXIES contract
// (PR-D follow-up, work_260512-i). Empty / "none" → nil keeps the silent
// default. "*" expands to dual-stack any. Comma lists are trimmed and
// emptied.
func TestTrustedProxiesFromEnv(t *testing.T) {
	cases := []struct {
		name string
		env  string
		want []string
	}{
		{name: "unset → nil", env: "", want: nil},
		{name: "none → nil (alias)", env: "none", want: nil},
		{name: "NONE → nil (case insensitive)", env: "NONE", want: nil},
		{name: "wildcard → dual-stack any", env: "*", want: []string{"0.0.0.0/0", "::/0"}},
		{name: "single CIDR", env: "10.0.0.0/8", want: []string{"10.0.0.0/8"}},
		{name: "comma list with whitespace", env: "10.0.0.0/8 , 192.168.1.5", want: []string{"10.0.0.0/8", "192.168.1.5"}},
		{name: "all empty entries → nil", env: " , , ", want: nil},
		{name: "leading + trailing whitespace", env: "  10.0.0.0/8  ", want: []string{"10.0.0.0/8"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("DEVHUB_TRUSTED_PROXIES", tc.env)
			got := trustedProxiesFromEnv()
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("trustedProxiesFromEnv() = %#v, want %#v", got, tc.want)
			}
		})
	}
}
