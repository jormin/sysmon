package main

import "testing"

func TestParseVersion(t *testing.T) {
	cases := []struct {
		in   string
		want []int
		ok   bool
	}{
		{"v0.0.1", []int{0, 0, 1}, true},
		{"0.0.1", []int{0, 0, 1}, true},
		{"v1.2.3", []int{1, 2, 3}, true},
		{" v2.10.0 ", []int{2, 10, 0}, true},
		{"dev", nil, false},
		{"v1.2", nil, false},
		{"", nil, false},
	}
	for _, c := range cases {
		got, err := parseVersion(c.in)
		if c.ok != (err == nil) {
			t.Errorf("parseVersion(%q) err=%v, want ok=%v", c.in, err, c.ok)
			continue
		}
		if c.ok {
			for i := range c.want {
				if got[i] != c.want[i] {
					t.Errorf("parseVersion(%q) = %v, want %v", c.in, got, c.want)
					break
				}
			}
		}
	}
}

func TestCompareVersions(t *testing.T) {
	older := []int{0, 0, 1}
	same := []int{0, 0, 1}
	newer := []int{0, 0, 2}
	muchNewer := []int{1, 0, 0}
	if compareVersions(older, same) != 0 {
		t.Errorf("equal versions should be 0")
	}
	if compareVersions(older, newer) != -1 {
		t.Errorf("older vs newer should be -1")
	}
	if compareVersions(newer, older) != 1 {
		t.Errorf("newer vs older should be 1")
	}
	if compareVersions(newer, muchNewer) != -1 {
		t.Errorf("0.0.2 vs 1.0.0 should be -1")
	}
}
