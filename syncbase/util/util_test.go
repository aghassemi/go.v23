// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util_test

import (
	"testing"

	"v.io/syncbase/v23/syncbase/util"
)

func TestValidName(t *testing.T) {
	tests := []struct {
		name string
		res  bool
	}{
		{"", false},
		{"*", false},
		{"a*", false},
		{"*a", false},
		{"a*b", false},
		{"/", false},
		{"a/", false},
		{"/a", false},
		{"a/b", false},
		{"a", true},
		{"aa", true},
	}
	for _, test := range tests {
		res := util.ValidName(test.name)
		if res != test.res {
			t.Errorf("%q: got %v, want %v", test.name, res, test.res)
		}
	}
}

func TestPrefixRange(t *testing.T) {
	tests := []struct {
		prefix string
		start  string
		limit  string
	}{
		{"", "", ""},
		{"a", "a", "b"},
		{"aa", "aa", "ab"},
		{"\xfe", "\xfe", "\xff"},
		{"a\xfe", "a\xfe", "a\xff"},
		{"aa\xfe", "aa\xfe", "aa\xff"},
		{"a\xff", "a\xff", "b"},
		{"aa\xff", "aa\xff", "ab"},
		{"a\xff\xff", "a\xff\xff", "b"},
		{"aa\xff\xff", "aa\xff\xff", "ab"},
		{"\xff", "\xff", ""},
		{"\xff\xff", "\xff\xff", ""},
	}
	for _, test := range tests {
		start, limit := util.PrefixRangeStart(test.prefix), util.PrefixRangeLimit(test.prefix)
		if start != test.start || limit != test.limit {
			t.Errorf("%q: got {%q, %q}, want {%q, %q}", test.prefix, start, limit, test.start, test.limit)
		}
	}
}

func TestIsPrefix(t *testing.T) {
	tests := []struct {
		isPrefix bool
		start  string
		limit  string
	}{
		{true, "", ""},
		{true, "a", "b"},
		{true, "aa", "ab"},
		{true, "\xfe", "\xff"},
		{true, "a\xfe", "a\xff"},
		{true, "aa\xfe", "aa\xff"},
		{true, "a\xff", "b"},
		{true, "aa\xff", "ab"},
		{true, "a\xff\xff", "b"},
		{true, "aa\xff\xff", "ab"},
		{true, "\xff", ""},
		{true, "\xff\xff", ""},

		{false, "", "\x00"},
		{false, "a", "aa"},
		{false, "aa", "aa"},
		{false, "\xfe", "\x00"},
		{false, "a\xfe", "b\xfe"},
		{false, "aa\xfe", "aa\x00"},
		{false, "a\xff", "b\x00"},
		{false, "aa\xff", "ab\x00"},
		{false, "a\xff\xff", "a\xff\xff\xff"},
		{false, "aa\xff\xff", "a"},
		{false, "\xff", "\x00"},
	}
	for _, test := range tests {
		result := util.IsPrefix(test.start, test.limit)
		if result != test.isPrefix {
			t.Errorf("%q, %q: got %v, want %v", test.start, test.limit, result, test.isPrefix)
		}
	}
}
