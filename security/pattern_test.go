package security

import (
	"reflect"
	"testing"
)

func TestMatchedBy(t *testing.T) {
	type v []string
	tests := []struct {
		Pattern      BlessingPattern
		Matches      v
		DoesNotMatch v
	}{
		{
			Pattern:      "",
			DoesNotMatch: v{"", "ann", "bob", "ann/friend"},
		},
		{
			Pattern:      "$",
			DoesNotMatch: v{"", "$", "ann", "bob", "ann/friend"},
		},
		{
			Pattern: AllPrincipals,
			Matches: v{"", "ann", "bob", "ann/friend"},
		},
		{
			Pattern:      "ann/$/$",
			DoesNotMatch: v{"", "ann", "bob", "ann/friend", "ann/friend/spouse"},
		},
		{
			Pattern:      "ann",
			Matches:      v{"ann", "ann/friend", "ann/enemy"},
			DoesNotMatch: v{"", "bob", "bob/ann"},
		},
		{
			Pattern:      "ann/friend",
			Matches:      v{"ann/friend", "ann/friend/spouse"},
			DoesNotMatch: v{"", "ann", "ann/enemy", "bob", "bob/ann"},
		},
		{
			Pattern:      "ann/friend/$",
			Matches:      v{"ann/friend"},
			DoesNotMatch: v{"", "ann", "ann/enemy", "ann/friend/spouse", "bob", "bob/friend", "bob/ann"},
		},
	}
	for _, test := range tests {
		// All combinations of test.Matches should match.
		for i := 0; i < len(test.Matches); i++ {
			for j := i + 1; j < len(test.Matches); j++ {
				args := []string(test.Matches[i:j])
				if !test.Pattern.MatchedBy(args...) {
					t.Errorf("%q.MatchedBy(%v) returned false", test.Pattern, args)
				}
			}
		}
		// All combinations of test.DoesNotMatch should not match.
		for i := 0; i < len(test.DoesNotMatch); i++ {
			for j := i + 1; j < len(test.DoesNotMatch); j++ {
				args := []string(test.DoesNotMatch[i:j])
				if test.Pattern.MatchedBy(args...) {
					t.Errorf("%q.MatchedBy(%v) returned true", test.Pattern, args)
				}
			}
		}
	}
}

func TestMatchedByCornerCases(t *testing.T) {
	if !AllPrincipals.MatchedBy() {
		t.Errorf("%q.MatchedBy() failed", AllPrincipals)
	}
	if NoExtension.MatchedBy() {
		t.Errorf("%q.MatchedBy() returned true", NoExtension)
	}
	if BlessingPattern("ann/$").MatchedBy() {
		t.Errorf("%q.MatchedBy() returned true", "ann/$")
	}
	if BlessingPattern("ann").MatchedBy() {
		t.Errorf("%q.MatchedBy() returned true", "ann")
	}
	if !AllPrincipals.MatchedBy("") {
		t.Errorf("%q.MatchedBy(%q) failed", AllPrincipals, "")
	}
	if NoExtension.MatchedBy("") {
		t.Errorf("%q.MatchedBy() returned true", NoExtension)
	}
	if BlessingPattern("ann/$").MatchedBy("") {
		t.Errorf("%q.MatchedBy(%q) returned true", "ann/$", "")
	}
	if BlessingPattern("ann").MatchedBy("") {
		t.Errorf("%q.MatchedBy(%q) returned true", "ann", "")
	}
}

func TestIsValid(t *testing.T) {
	var (
		valid   = []BlessingPattern{AllPrincipals, "alice", "al$ice", "alice/$", "alice.jones/$", "alice@google/$", "veyron/alice@google/$", "veyron/alice@google/bob/$", "alice", "alice/bob"}
		invalid = []BlessingPattern{"", "alice...", "...alice", "alice...bob", "/alice", "alice/", "...alice/bob", "alice.../bob", "alice/.../bob", "alice/$/bob", "alice/$/$", "alice/.../$", "alice/..."}
	)
	for _, p := range valid {
		if !p.IsValid() {
			t.Errorf("%q.IsValid() returned false", p)
		}
	}
	for _, p := range invalid {
		if p.IsValid() {
			t.Errorf("%q.IsValid() returned true", p)
		}
	}
}

func TestMakeNonExtendable(t *testing.T) {
	tests := []struct{ before, after BlessingPattern }{
		{"", "$"},
		{"$", "$"},
		{"a", "a/$"},
		{"a/$", "a/$"},
	}
	for _, test := range tests {
		if got, want := test.before.MakeNonExtendable(), test.after; got != want {
			t.Errorf("%q.MakeNonExtendable(): Got %q, want %q", test.before, got, want)
		}
	}
}

func TestPrefixPatterns(t *testing.T) {
	tests := []struct {
		pattern  BlessingPattern
		prefixes []BlessingPattern
	}{
		{"$", []BlessingPattern{"$"}},
		{"ann", []BlessingPattern{"ann"}},
		{"ann/$", []BlessingPattern{"ann/$"}},
		{"ann/friend", []BlessingPattern{"ann/$", "ann/friend"}},
		{"ann/friend/$", []BlessingPattern{"ann/$", "ann/friend/$"}},
	}
	for _, test := range tests {
		if got, want := test.pattern.PrefixPatterns(), test.prefixes; !reflect.DeepEqual(got, want) {
			t.Fatalf("%q.PrefixPatterns(): got %q, want %q", test.pattern, got, want)
		}
	}
}

func TestSplitPatternName(t *testing.T) {
	const notset = ""
	cases := []struct {
		input  string
		server BlessingPattern
		name   string
	}{
		{"", notset, ""},
		{"/", notset, "/"},
		{"__(foo/bar)", "foo/bar", ""},
		{"/__(foo/bar)", "foo/bar", "/"},
		{"/__(x/y)", "x/y", "/"},
		{"/__(x/y)/a/b", notset, "/__(x/y)/a/b"},
		{"a/__(foo)", "foo", "a"},
		{"/__(foo)a", notset, "/__(foo)a"},
		{"/__(foo)a/__(bar)", "bar", "/__(foo)a"},
		{"a/b", notset, "a/b"},
		{"a/b/__(foo)", "foo", "a/b"},
		{"/a/b", notset, "/a/b"},
		{"/__(foo)a/b", notset, "/__(foo)a/b"},
		{"/a/b/__(bar)", "bar", "/a/b"},
		{"/__(foo)a/b/__(bar)", "bar", "/__(foo)a/b"},
		{"/a/b__(foo)", notset, "/a/b__(foo)"},
		{"/a/b/__(foo)c", notset, "/a/b/__(foo)c"},
		{"/[01:02::]:444", notset, "/[01:02::]:444"},
		{"/__(foo)[01:02::]:444", notset, "/__(foo)[01:02::]:444"},
		{"/[01:02::]:444/foo", notset, "/[01:02::]:444/foo"},
		{"/__(a)[01:02::]:444/foo", notset, "/__(a)[01:02::]:444/foo"},
		{"/[01:02::]:444/foo/__(b)", "b", "/[01:02::]:444/foo"},
		{"/__(c)[01:02::]:444/foo/__(d)", "d", "/__(c)[01:02::]:444/foo"},
	}
	for _, c := range cases {
		server, name := SplitPatternName(c.input)
		if server != c.server {
			t.Errorf("%q: unexpected server pattern: %q not %q", c.input, server, c.server)
		}
		if name != c.name {
			t.Errorf("%q: unexpected name: %q not %q", c.input, name, c.name)
		}
	}
}
