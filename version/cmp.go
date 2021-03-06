// Copyright (c) 2020 Tailscale Inc & AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package version

import (
	"strconv"
	"strings"
)

// AtLeast returns whether version is at least the specified minimum
// version.
//
// Version comparison in Tailscale is a little complex, because we
// switched "styles" a few times, and additionally have a completely
// separate track of version numbers for OSS-only builds.
//
// AtLeast acts conservatively, returning true only if it's certain
// that version is at least minimum. As a result, it can produce false
// negatives, for example when an OSS build supports a given feature,
// but AtLeast is called with an official release number as the
// minimum
//
// version and minimum can both be either an official Tailscale
// version numbers (major.minor.patch-extracommits-extrastring), or an
// OSS build datestamp (date.YYYYMMDD). For Tailscale version numbers,
// AtLeast also accepts a prefix of a full version, in which case all
// missing fields are assumed to be zero.
func AtLeast(version string, minimum string) bool {
	v, ok := parse(version)
	if !ok {
		return false
	}
	m, ok := parse(minimum)
	if !ok {
		return false
	}

	switch {
	case v.Datestamp != 0 && m.Datestamp == 0:
		// OSS version vs. Tailscale version
		return false
	case v.Datestamp == 0 && m.Datestamp != 0:
		// Tailscale version vs. OSS version
		return false
	case v.Datestamp != 0:
		// OSS version vs. OSS version
		return v.Datestamp >= m.Datestamp
	case v.Major == m.Major && v.Minor == m.Minor && v.Patch == m.Patch && v.ExtraCommits == m.ExtraCommits:
		// Exactly equal Tailscale versions
		return true
	case v.Major != m.Major:
		return v.Major > m.Major
	case v.Minor != m.Minor:
		return v.Minor > m.Minor
	case v.Patch != m.Patch:
		return v.Patch > m.Patch
	default:
		return v.ExtraCommits > m.ExtraCommits
	}
}

type parsed struct {
	Major, Minor, Patch, ExtraCommits int // for Tailscale version e.g. e.g. "0.99.1-20"
	Datestamp                         int // for OSS version e.g. "date.20200612"
}

func parse(version string) (parsed, bool) {
	if strings.HasPrefix(version, "date.") {
		stamp, err := strconv.Atoi(version[5:])
		if err != nil {
			return parsed{}, false
		}
		return parsed{Datestamp: stamp}, true
	}

	var ret parsed

	major, rest, err := splitNumericPrefix(version)
	if err != nil {
		return parsed{}, false
	}
	ret.Major = major
	if rest == "" {
		return ret, true
	}

	ret.Minor, rest, err = splitNumericPrefix(rest[1:])
	if err != nil {
		return parsed{}, false
	}
	if rest == "" {
		return ret, true
	}

	// Optional patch version, if the next separator is a dot.
	if rest[0] == '.' {
		ret.Patch, rest, err = splitNumericPrefix(rest[1:])
		if err != nil {
			return parsed{}, false
		}
		if rest == "" {
			return ret, true
		}
	}

	// Optional extraCommits, if the next bit can be completely
	// consumed as an integer.
	if rest[0] != '-' {
		return parsed{}, false
	}

	var trailer string
	ret.ExtraCommits, trailer, err = splitNumericPrefix(rest[1:])
	if err != nil || (len(trailer) > 0 && trailer[0] != '-') {
		// rest was probably the string trailer, ignore it.
		ret.ExtraCommits = 0
	}
	return ret, true
}

func splitNumericPrefix(s string) (int, string, error) {
	for i, r := range s {
		if r >= '0' && r <= '9' {
			continue
		}
		ret, err := strconv.Atoi(s[:i])
		if err != nil {
			return 0, "", err
		}
		return ret, s[i:], nil
	}

	ret, err := strconv.Atoi(s)
	if err != nil {
		return 0, "", err
	}
	return ret, "", nil
}
