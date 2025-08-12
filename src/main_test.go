package main

import (
	"strings"
	"testing"
)

func TestExtractDomain(t *testing.T) {
	cases := []struct{ in, want string }{
		{"http://example.com/path", "example.com"},
		{"https://example.com:443/", "example.com"},
		{"example.org", "example.org"},
		{"example.org:8080/foo", "example.org"},
	}
	for _, c := range cases {
		if got := extractDomain(c.in); got != c.want {
			t.Fatalf("extractDomain(%q)=%q want %q", c.in, got, c.want)
		}
	}
}

func TestBuildAccessLogSummaryFromLog(t *testing.T) {
	log := "" +
		"1712175100.000 WL GET 200 example.com example.com:80\n" +
		"1712175101.000 BL GET 200 blocked.com blocked.com:443\n" +
		"1712175102.000 WL GET 200 example.com example.com:80\n"
	whitelist := "example.com\n# comment\n"
	blacklist := "blocked.com\n"
	summary := buildAccessLogSummaryFromLog(log, whitelist, blacklist)
	if !containsAll(summary, []string{"example.com", "blocked.com", "✅", "❌"}) {
		t.Fatalf("summary missing expected tokens:\n%s", summary)
	}
}

func TestIsDomainLikeRejectsNumeric(t *testing.T) {
	bad := []string{"1754845426.256", "123.456", "10.20.30.40"}
	for _, v := range bad {
		if isDomainLike(v) {
			t.Fatalf("expected isDomainLike(%q) == false", v)
		}
	}
	good := []string{"example.com", "sub.domain-1.org", "a.b"}
	for _, v := range good {
		if !isDomainLike(v) {
			t.Fatalf("expected isDomainLike(%q) == true", v)
		}
	}
}

func containsAll(s string, subs []string) bool {
	for _, sub := range subs {
		if !strings.Contains(s, sub) {
			return false
		}
	}
	return true
}
