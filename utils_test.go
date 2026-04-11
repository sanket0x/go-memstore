package memstore

import "testing"

// Empty pattern and wildcard-only pattern tests
func TestMatchPattern_EmptyPatternMatchesEverything(t *testing.T) {
	result := matchPattern("anything", "")
	if result != true {
		t.Errorf("matchPattern(%q, %q) = %v, expected %v", "anything", "", result, true)
	}
}

func TestMatchPattern_SingleWildcardMatchesEverything(t *testing.T) {
	result := matchPattern("anything", "*")
	if result != true {
		t.Errorf("matchPattern(%q, %q) = %v, expected %v", "anything", "*", result, true)
	}
}

func TestMatchPattern_EmptyStringWithEmptyPattern(t *testing.T) {
	result := matchPattern("", "")
	if result != true {
		t.Errorf("matchPattern(%q, %q) = %v, expected %v", "", "", result, true)
	}
}

func TestMatchPattern_EmptyStringWithWildcard(t *testing.T) {
	result := matchPattern("", "*")
	if result != true {
		t.Errorf("matchPattern(%q, %q) = %v, expected %v", "", "*", result, true)
	}
}

// Exact match tests
func TestMatchPattern_ExactMatchSuccess(t *testing.T) {
	result := matchPattern("hello", "hello")
	if result != true {
		t.Errorf("matchPattern(%q, %q) = %v, expected %v", "hello", "hello", result, true)
	}
}

func TestMatchPattern_ExactMatchFailure(t *testing.T) {
	result := matchPattern("hello", "world")
	if result != false {
		t.Errorf("matchPattern(%q, %q) = %v, expected %v", "hello", "world", result, false)
	}
}

func TestMatchPattern_ExactMatchWithEmptyString(t *testing.T) {
	result := matchPattern("", "hello")
	if result != false {
		t.Errorf("matchPattern(%q, %q) = %v, expected %v", "", "hello", result, false)
	}
}

// Prefix match tests
func TestMatchPattern_PrefixMatchSuccess(t *testing.T) {
	result := matchPattern("hello world", "hello*")
	if result != true {
		t.Errorf("matchPattern(%q, %q) = %v, expected %v", "hello world", "hello*", result, true)
	}
}

func TestMatchPattern_PrefixMatchFailure(t *testing.T) {
	result := matchPattern("world hello", "hello*")
	if result != false {
		t.Errorf("matchPattern(%q, %q) = %v, expected %v", "world hello", "hello*", result, false)
	}
}

func TestMatchPattern_PrefixMatchWithExactString(t *testing.T) {
	result := matchPattern("hello", "hello*")
	if result != true {
		t.Errorf("matchPattern(%q, %q) = %v, expected %v", "hello", "hello*", result, true)
	}
}

// Suffix match tests
func TestMatchPattern_SuffixMatchSuccess(t *testing.T) {
	result := matchPattern("hello world", "*world")
	if result != true {
		t.Errorf("matchPattern(%q, %q) = %v, expected %v", "hello world", "*world", result, true)
	}
}

func TestMatchPattern_SuffixMatchFailure(t *testing.T) {
	result := matchPattern("hello world test", "*world")
	if result != false {
		t.Errorf("matchPattern(%q, %q) = %v, expected %v", "hello world test", "*world", result, false)
	}
}

func TestMatchPattern_SuffixMatchWithExactString(t *testing.T) {
	result := matchPattern("world", "*world")
	if result != true {
		t.Errorf("matchPattern(%q, %q) = %v, expected %v", "world", "*world", result, true)
	}
}

// Complex patterns with multiple wildcards
func TestMatchPattern_MiddleWildcardSuccess(t *testing.T) {
	result := matchPattern("hello beautiful world", "hello*world")
	if result != true {
		t.Errorf("matchPattern(%q, %q) = %v, expected %v", "hello beautiful world", "hello*world", result, true)
	}
}

func TestMatchPattern_MiddleWildcardFailure(t *testing.T) {
	result := matchPattern("hello world beautiful", "hello*world")
	if result != false {
		t.Errorf("matchPattern(%q, %q) = %v, expected %v", "hello world beautiful", "hello*world", result, false)
	}
}

func TestMatchPattern_MultipleWildcardsSuccess(t *testing.T) {
	result := matchPattern("prefix middle1 middle2 suffix", "prefix*middle1*suffix")
	if result != true {
		t.Errorf("matchPattern(%q, %q) = %v, expected %v", "prefix middle1 middle2 suffix", "prefix*middle1*suffix", result, true)
	}
}

func TestMatchPattern_MultipleWildcardsFailureMissingMiddle(t *testing.T) {
	result := matchPattern("prefix middle2 suffix", "prefix*middle1*suffix")
	if result != false {
		t.Errorf("matchPattern(%q, %q) = %v, expected %v", "prefix middle2 suffix", "prefix*middle1*suffix", result, false)
	}
}

func TestMatchPattern_MultipleWildcardsFailureWrongOrder(t *testing.T) {
	result := matchPattern("prefix middle2 middle1 suffix", "prefix*middle1*middle2*suffix")
	if result != false {
		t.Errorf("matchPattern(%q, %q) = %v, expected %v", "prefix middle2 middle1 suffix", "prefix*middle1*middle2*suffix", result, false)
	}
}

// Edge cases
func TestMatchPattern_ConsecutiveWildcards(t *testing.T) {
	result := matchPattern("hello world", "hello**world")
	if result != true {
		t.Errorf("matchPattern(%q, %q) = %v, expected %v", "hello world", "hello**world", result, true)
	}
}

func TestMatchPattern_PatternLongerThanInput(t *testing.T) {
	result := matchPattern("hi", "hello")
	if result != false {
		t.Errorf("matchPattern(%q, %q) = %v, expected %v", "hi", "hello", result, false)
	}
}

func TestMatchPattern_WildcardWithEmptyParts(t *testing.T) {
	result := matchPattern("test", "*test*")
	if result != true {
		t.Errorf("matchPattern(%q, %q) = %v, expected %v", "test", "*test*", result, true)
	}
}

func TestMatchPattern_OnlyWildcards(t *testing.T) {
	result := matchPattern("anything", "***")
	if result != true {
		t.Errorf("matchPattern(%q, %q) = %v, expected %v", "anything", "***", result, true)
	}
}

func TestMatchPattern_ComplexRealWorldExampleSuccess(t *testing.T) {
	result := matchPattern("user:123:profile:settings", "user:*:profile:*")
	if result != true {
		t.Errorf("matchPattern(%q, %q) = %v, expected %v", "user:123:profile:settings", "user:*:profile:*", result, true)
	}
}

func TestMatchPattern_ComplexRealWorldExampleFailure(t *testing.T) {
	result := matchPattern("user:123:account:settings", "user:*:profile:*")
	if result != false {
		t.Errorf("matchPattern(%q, %q) = %v, expected %v", "user:123:account:settings", "user:*:profile:*", result, false)
	}
}

// Special characters and edge cases
func TestMatchPattern_PatternWithSpecialCharacters(t *testing.T) {
	result := matchPattern("test@email.com", "test@*")
	if result != true {
		t.Errorf("matchPattern(%q, %q) = %v, expected %v", "test@email.com", "test@*", result, true)
	}
}

func TestMatchPattern_InputWithSpecialCharactersNoMatch(t *testing.T) {
	result := matchPattern("test@email.com", "user@*")
	if result != false {
		t.Errorf("matchPattern(%q, %q) = %v, expected %v", "test@email.com", "user@*", result, false)
	}
}

func TestMatchPattern_SingleCharacterInputsAndPatterns(t *testing.T) {
	result := matchPattern("a", "a")
	if result != true {
		t.Errorf("matchPattern(%q, %q) = %v, expected %v", "a", "a", result, true)
	}
}

func TestMatchPattern_SingleCharacterWithWildcard(t *testing.T) {
	result := matchPattern("a", "*a")
	if result != true {
		t.Errorf("matchPattern(%q, %q) = %v, expected %v", "a", "*a", result, true)
	}
}

func TestMatchPattern_WildcardInMiddleOfSingleChars(t *testing.T) {
	result := matchPattern("abc", "a*c")
	if result != true {
		t.Errorf("matchPattern(%q, %q) = %v, expected %v", "abc", "a*c", result, true)
	}
}

// Tests based on the function's documentation examples
func TestMatchPattern_WildcardMatchesEverything(t *testing.T) {
	if !matchPattern("anything", "*") {
		t.Error("'*' should match everything")
	}
	if !matchPattern("", "*") {
		t.Error("'*' should match empty string")
	}
}

func TestMatchPattern_PrefixExample(t *testing.T) {
	if !matchPattern("prefix123", "prefix*") {
		t.Error("'prefix*' should match 'prefix123'")
	}
	if matchPattern("123prefix", "prefix*") {
		t.Error("'prefix*' should not match '123prefix'")
	}
}

func TestMatchPattern_SuffixExample(t *testing.T) {
	if !matchPattern("123suffix", "*suffix") {
		t.Error("'*suffix' should match '123suffix'")
	}
	if matchPattern("suffix123", "*suffix") {
		t.Error("'*suffix' should not match 'suffix123'")
	}
}

func TestMatchPattern_MultiplePartsInOrderExample(t *testing.T) {
	if !matchPattern("pre123mid456suf", "pre*mid*suf") {
		t.Error("'pre*mid*suf' should match 'pre123mid456suf'")
	}
	if matchPattern("pre123suf456mid", "pre*mid*suf") {
		t.Error("'pre*mid*suf' should not match 'pre123suf456mid' (wrong order)")
	}
}

func TestMatchPattern_ExactExample(t *testing.T) {
	if !matchPattern("exact", "exact") {
		t.Error("'exact' should match 'exact'")
	}
	if matchPattern("exact123", "exact") {
		t.Error("'exact' should not match 'exact123'")
	}
}
