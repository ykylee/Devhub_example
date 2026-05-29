package integrationcaps

import (
	"testing"

	"github.com/devhub/backend-core/internal/domain"
)

func TestProviderHasCapability(t *testing.T) {
	cases := []struct {
		name     string
		caps     []string
		want     []string
		expected bool
	}{
		{name: "empty capabilities + no want", caps: nil, want: nil, expected: false},
		{name: "empty capabilities + one want", caps: nil, want: []string{"pull"}, expected: false},
		{name: "single capability + matching want", caps: []string{"pull"}, want: []string{"pull"}, expected: true},
		{name: "single capability + non-matching want", caps: []string{"webhook"}, want: []string{"pull"}, expected: false},
		{name: "single capability + multi want OR match", caps: []string{"pull"}, want: []string{"pull", "sync"}, expected: true},
		{name: "single capability + multi want OR via second", caps: []string{"sync"}, want: []string{"pull", "sync"}, expected: true},
		{name: "multi capabilities + match first want", caps: []string{"pull", "webhook"}, want: []string{"pull"}, expected: true},
		{name: "multi capabilities + match second want", caps: []string{"pull", "webhook"}, want: []string{"sync", "webhook"}, expected: true},
		{name: "multi capabilities + no match", caps: []string{"pull", "webhook"}, want: []string{"push", "sync"}, expected: false},
		{name: "no want provided returns false (vacuous OR)", caps: []string{"pull"}, want: nil, expected: false},
		{name: "duplicate capabilities are ignored", caps: []string{"pull", "pull", "pull"}, want: []string{"pull"}, expected: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := domain.IntegrationProvider{Capabilities: tc.caps}
			got := ProviderHasCapability(p, tc.want...)
			if got != tc.expected {
				t.Fatalf("ProviderHasCapability(caps=%v, want=%v) = %v, want %v",
					tc.caps, tc.want, got, tc.expected)
			}
		})
	}
}
