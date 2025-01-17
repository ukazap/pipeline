// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xos_test

import (
	"reflect"
	"strings"
	"testing"

	. "github.com/unchain/pipeline/pkg/xos"

	"github.com/stretchr/testify/require"
)

func TestName(t *testing.T) {
	x := "0123456789"

	require.Equal(t, "012", x[0:3])
	require.Equal(t, "3", x[3:4])
	require.Equal(t, "012", x[0:3])
}

// testGetenv gives us a controlled set of variables for testing Expand.
func testGetenv(s string) string {
	switch s {
	case "*":
		return "all the args"
	case "#":
		return "NARGS"
	case "$":
		return "PID"
	case "1":
		return "ARGUMENT1"
	case "home.home":
		return "/usr/gopher2"
	case "HOME":
		return "/usr/gopher"
	case "H":
		return "(Value of H)"
	case "home_1":
		return "/usr/foo"
	case "_":
		return "underscore"
	}
	return ""
}

var expandTests = []struct {
	in, out string
}{
	{"", ""},
	{"$*", "all the args"},
	{"$$", "PID"},
	{"${*}", "all the args"},
	{"$1", "ARGUMENT1"},
	{"${1}", "ARGUMENT1"},
	{"now is the time", "now is the time"},
	{"$HOME", "/usr/gopher"},
	{"$home_1", "/usr/foo"},
	{"${HOME}", "/usr/gopher"},
	{"${H}OME", "(Value of H)OME"},
	{"A$$$#$1$H$home_1*B", "APIDNARGSARGUMENT1(Value of H)/usr/foo*B"},
	{"start$+middle$^end$", "start$+middle$^end$"},
	{"mixed$|bag$$$", "mixed$|bagPID$"},
	{"$", "$"},
	{"$}", "$}"},
	{"${", ""},  // invalid syntax; eat up the characters
	{"${}", ""}, // invalid syntax; eat up the characters
}

var expandTests2 = []struct {
	in, out string
}{
	{"", ""},
	{"$.*", "all the args"},
	{"$.$", "PID"},
	{"$.{*}", "all the args"},
	{"$.1", "ARGUMENT1"},
	{"$.{1}", "ARGUMENT1"},
	{"now is the time", "now is the time"},
	{"$.HOME", "/usr/gopher"},
	{"$.home_1", "/usr/foo"},
	{"$.{HOME}", "/usr/gopher"},
	{"$.{H}OME", "(Value of H)OME"},
	{"A$.$$.#$.1$.H$.home_1*B", "APIDNARGSARGUMENT1(Value of H)/usr/foo*B"},
	{"start$.+middle$.^end$.", "start$.+middle$.^end$."},
	{"mixed$.|bag$.$$.", "mixed$.|bagPID$."},
	{"$.", "$."},
	{"$.}", "$.}"},
	{"$.{", ""},  // invalid syntax; eat up the characters
	{"$.{}", ""}, // invalid syntax; eat up the characters
}

var expandTests3 = []struct {
	in, out string
}{
	{"", ""},
	{"$.#blabla*", "all the args"},
	{"$.#blabla$", "PID"},
	{"$.#blabla{*}", "all the args"},
	{"$.#blabla1", "ARGUMENT1"},
	{"$.#blabla{1}", "ARGUMENT1"},
	{"now is the time", "now is the time"},
	{"$.#blablaHOME", "/usr/gopher"},
	{"$.#blablahome_1", "/usr/foo"},
	{"$.#blabla{HOME}", "/usr/gopher"},
	{"$.#blabla{H}OME", "(Value of H)OME"},
	{"A$.#blabla$$.#blabla#$.#blabla1$.#blablaH$.#blablahome_1*B", "APIDNARGSARGUMENT1(Value of H)/usr/foo*B"},
	{"start$.#blabla+middle$.#blabla^end$.#blabla", "start$.#blabla+middle$.#blabla^end$.#blabla"},
	{"mixed$.#blabla|bag$.#blabla$$.#blabla", "mixed$.#blabla|bagPID$.#blabla"},
	{"$.#blabla", "$.#blabla"},
	{"$.#blabla}", "$.#blabla}"},
	{"$.#blabla{", ""},  // invalid syntax; eat up the characters
	{"$.#blabla{}", ""}, // invalid syntax; eat up the characters
}

var multiExpandTests = []struct {
	expanders []*Expander
	cases     []struct {
		in, out string
	}
}{
	{
		[]*Expander{{"$", testGetenv}},
		expandTests,
	},
	{
		[]*Expander{{"$.", testGetenv}},
		expandTests2,
	},
	{
		[]*Expander{{"$.#blabla", testGetenv}},
		expandTests3,
	},
	{
		[]*Expander{
			{"$.#blabla", testGetenv},
			{"$.", testGetenv},
			{"$", func(s string) string {
				if s == "$" {
					return "$"
				}

				return testGetenv(s)
			}}},
		[]struct {
			in, out string
		}{
			{"", ""},
			{"$somethingnonexisting$.#blabla*", "all the args"},
			{"$HOME$.HOME$.#blabla$", "/usr/gopher/usr/gopherPID"},
			{"$.#blabla{*}", "all the args"},
			{"$.#blabla1", "ARGUMENT1"},
			{"$.#blabla{1}", "ARGUMENT1"},
			{"now is the time", "now is the time"},
			{"$.#blablaHOME", "/usr/gopher"},
			{"$.#blablahome_1", "/usr/foo"},
			{"$.#blabla{HOME}", "/usr/gopher"},
			{"$.#blabla{H}OME", "(Value of H)OME"},
			{"A$.#blabla$$.#blabla#$.#blabla1$.#blablaH$.#blablahome_1*B", "APIDNARGSARGUMENT1(Value of H)/usr/foo*B"},
			{"start$.#blabla+middle$.#blabla^end$.#blabla", "start$.#blabla+middle$.#blabla^end$.#blabla"},
			{"mixed$.#blabla|bag$.#blabla$$.#blabla", "mixed$.#blabla|bagPID$.#blabla"},
			{"$.#blabla", "$.#blabla"},
			{"$.#blabla}", "$.#blabla}"},
			{"$.#blabla{", ""},                                // invalid syntax; eat up the characters
			{"$.#blabla{}", ""},                               // invalid syntax; eat up the characters
			{"$$.#blabla{home}", "$.#blabla{home}"},           // invalid syntax; eat up the characters
			{"$$.#blabla{home.home}", "$.#blabla{home.home}"}, // invalid syntax; eat up the characters

		},
	},
}

func TestMultiExpand(t *testing.T) {
	for _, test := range multiExpandTests {
		for _, tc := range test.cases {
			result := MultiExpand(tc.in, test.expanders)
			if result != tc.out {
				t.Errorf("Expand(%q)=%q; expected %q", tc.in, result, tc.out)
			}
		}
	}
}

func TestMultiExpandOne(t *testing.T) {
	tc := struct {
		expanders []*Expander
		in, out   string
	}{
		[]*Expander{
			{"$.#blabla", func(s string) string {
				return testGetenv(s)
			}},
			{"$.", func(s string) string {
				return testGetenv(s)
			}},
			{"$", func(s string) string {
				if s == "$" {
					return "$"
				}

				return testGetenv(s)
			}}},
		"blaz $.{home.home}", "blaz /usr/gopher2", // invalid syntax; eat up the characters
	}

	//tc.in = strings.Replace(tc.in, "$$", "${JANUS_DOLLAR_ESCAPE_SIGN}", -1)

	result := MultiExpand(tc.in, tc.expanders)
	if result != tc.out {
		t.Errorf("Expand(%q)=%q; expected %q", tc.in, result, tc.out)
	}
}

func TestExpand(t *testing.T) {
	for _, test := range expandTests {
		result := Expand(test.in, testGetenv)
		if result != test.out {
			t.Errorf("Expand(%q)=%q; expected %q", test.in, result, test.out)
		}
	}
}

var global interface{}

func BenchmarkExpand(b *testing.B) {
	b.Run("noop", func(b *testing.B) {
		var s string
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			s = Expand("tick tick tick tick", func(string) string { return "" })
		}
		global = s
	})
	b.Run("multiple", func(b *testing.B) {
		var s string
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			s = Expand("$a $a $a $a", func(string) string { return "boom" })
		}
		global = s
	})
}

func TestConsistentEnviron(t *testing.T) {
	e0 := Environ()
	for i := 0; i < 10; i++ {
		e1 := Environ()
		if !reflect.DeepEqual(e0, e1) {
			t.Fatalf("environment changed")
		}
	}
}

func TestUnsetenv(t *testing.T) {
	const testKey = "GO_TEST_UNSETENV"
	set := func() bool {
		prefix := testKey + "="
		for _, key := range Environ() {
			if strings.HasPrefix(key, prefix) {
				return true
			}
		}
		return false
	}
	if err := Setenv(testKey, "1"); err != nil {
		t.Fatalf("Setenv: %v", err)
	}
	if !set() {
		t.Error("Setenv didn't set TestUnsetenv")
	}
	if err := Unsetenv(testKey); err != nil {
		t.Fatalf("Unsetenv: %v", err)
	}
	if set() {
		t.Fatal("Unsetenv didn't clear TestUnsetenv")
	}
}

func TestClearenv(t *testing.T) {
	const testKey = "GO_TEST_CLEARENV"
	const testValue = "1"

	// reset env
	defer func(origEnv []string) {
		for _, pair := range origEnv {
			// Environment variables on Windows can begin with =
			// https://blogs.msdn.com/b/oldnewthing/archive/2010/05/06/10008132.aspx
			i := strings.Index(pair[1:], "=") + 1
			if err := Setenv(pair[:i], pair[i+1:]); err != nil {
				t.Errorf("Setenv(%q, %q) failed during reset: %v", pair[:i], pair[i+1:], err)
			}
		}
	}(Environ())

	if err := Setenv(testKey, testValue); err != nil {
		t.Fatalf("Setenv(%q, %q) failed: %v", testKey, testValue, err)
	}
	if _, ok := LookupEnv(testKey); !ok {
		t.Errorf("Setenv(%q, %q) didn't set $%s", testKey, testValue, testKey)
	}
	Clearenv()
	if val, ok := LookupEnv(testKey); ok {
		t.Errorf("Clearenv() didn't clear $%s, remained with value %q", testKey, val)
	}
}

func TestLookupEnv(t *testing.T) {
	const smallpox = "SMALLPOX"      // No one has smallpox.
	value, ok := LookupEnv(smallpox) // Should not exist.
	if ok || value != "" {
		t.Fatalf("%s=%q", smallpox, value)
	}
	defer Unsetenv(smallpox)
	err := Setenv(smallpox, "virus")
	if err != nil {
		t.Fatalf("failed to release smallpox virus")
	}
	_, ok = LookupEnv(smallpox)
	if !ok {
		t.Errorf("smallpox release failed; world remains safe but LookupEnv is broken")
	}
}
