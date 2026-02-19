package contentassist

import (
	"testing"
)

func TestSuggestUsersEmptyPrefixReturnsTopFive(t *testing.T) {
	SetUserCache([]string{"alice", "bob", "carol", "dave", "eve", "frank"})

	results := SuggestUsers("")
	if len(results) != 5 {
		t.Errorf("expected 5 results for empty prefix, got %d", len(results))
	}
}

func TestSuggestUsersMatchesPrefix(t *testing.T) {
	SetUserCache([]string{"alice", "albert", "bob", "carol"})

	results := SuggestUsers("al")
	if len(results) != 2 {
		t.Errorf("expected 2 results for prefix 'al', got %d: %v", len(results), results)
	}
	if results[0] != "alice" || results[1] != "albert" {
		t.Errorf("unexpected results: %v", results)
	}
}

func TestSuggestUsersCaseInsensitive(t *testing.T) {
	SetUserCache([]string{"Alice", "ALBERT", "bob"})

	results := SuggestUsers("AL")
	if len(results) != 2 {
		t.Errorf("expected 2 results for case-insensitive prefix 'AL', got %d: %v", len(results), results)
	}
}

func TestSuggestUsersNoMatch(t *testing.T) {
	SetUserCache([]string{"alice", "bob"})

	results := SuggestUsers("xyz")
	if len(results) != 0 {
		t.Errorf("expected 0 results for non-matching prefix, got %d", len(results))
	}
}

func TestSuggestUsersCapsAtFive(t *testing.T) {
	SetUserCache([]string{"aa", "ab", "ac", "ad", "ae", "af", "ag"})

	results := SuggestUsers("a")
	if len(results) != 5 {
		t.Errorf("expected results capped at 5, got %d", len(results))
	}
}

func TestSuggestUsersEmptyCache(t *testing.T) {
	SetUserCache([]string{})

	results := SuggestUsers("")
	if len(results) != 0 {
		t.Errorf("expected 0 results for empty cache, got %d", len(results))
	}
}

func TestSetUserCacheReplaces(t *testing.T) {
	SetUserCache([]string{"alice"})
	SetUserCache([]string{"bob", "carol"})

	results := SuggestUsers("ali")
	if len(results) != 0 {
		t.Errorf("expected old cache to be replaced, got %v", results)
	}

	results2 := SuggestUsers("b")
	if len(results2) != 1 || results2[0] != "bob" {
		t.Errorf("expected new cache to contain 'bob', got %v", results2)
	}
}
