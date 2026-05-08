package config

import "testing"

func TestValidate(t *testing.T) {
	cases := []struct {
		name        string
		cfg         Config
		hasVerifier bool
		wantErr     bool
	}{
		{"dev: no verifier OK", Config{Env: ""}, false, false},
		{"dev: verifier OK", Config{Env: ""}, true, false},
		{"dev: dev fallback OK", Config{Env: "", AuthDevFallback: true}, false, false},
		{"dev (explicit): combinations OK", Config{Env: "dev", AuthDevFallback: true}, false, false},
		{"prod + verifier OK", Config{Env: "prod"}, true, false},
		{"prod + no verifier rejected", Config{Env: "prod"}, false, true},
		{"prod + verifier + dev fallback rejected", Config{Env: "prod", AuthDevFallback: true}, true, true},
		{"prod + no verifier + dev fallback rejected", Config{Env: "prod", AuthDevFallback: true}, false, true},
		{"prod uppercase normalized: no verifier rejected", Config{Env: "PROD"}, false, true},
		{"prod with whitespace normalized: verifier OK", Config{Env: " prod "}, true, false},
		{"prod with whitespace + dev fallback rejected", Config{Env: " Prod ", AuthDevFallback: true}, true, true},
	}
	for _, tc := range cases {
		err := tc.cfg.Validate(tc.hasVerifier)
		if tc.wantErr && err == nil {
			t.Errorf("%s: expected error, got nil", tc.name)
		}
		if !tc.wantErr && err != nil {
			t.Errorf("%s: expected no error, got %v", tc.name, err)
		}
	}
}
