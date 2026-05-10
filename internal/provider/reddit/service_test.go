package reddit

import (
	"strconv"
	"testing"
)

func TestNormalizeCustomAudienceNameQueryParam(t *testing.T) {
	t.Parallel()
	cases := []struct {
		raw     string
		want    string
		wantErr bool
	}{
		{"", "", false},
		{"  ", "", false},
		{"foo", "=@foo", false},
		{"testaudience", "=@testaudience", false},
		{"=@manual", "=@manual", false},
		{"==Exact", "==Exact", false},
		{"does-not-exist", "", true},
	}
	for _, tc := range cases {
		t.Run(strconv.Quote(tc.raw), func(t *testing.T) {
			t.Parallel()
			got, err := normalizeCustomAudienceNameQueryParam(tc.raw)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}
