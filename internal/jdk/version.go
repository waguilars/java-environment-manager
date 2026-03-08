package jdk

import (
	"regexp"
	"strconv"
)

// ParseVersion parses a JDK version string into its components
func ParseVersion(version string) (*Version, error) {
	// Pattern to match standard JDK version formats:
	// - 21.0.2
	// - 17.0.10+13
	// - 11.0.2
	// - 8u292
	pattern := `^(\d+)(?:\.(\d+))?(?:\.(\d+))?(?:\+(\d+))?$`
	re := regexp.MustCompile(pattern)

	matches := re.FindStringSubmatch(version)
	if matches == nil {
		return nil, &ParseVersionError{Version: version, Message: "invalid version format"}
	}

	v := &Version{
		Major: parseInt(matches[1]),
		Minor: parseIntDefault(matches[2], 0),
		Patch: parseIntDefault(matches[3], 0),
		Build: parseIntDefault(matches[4], 0),
		Raw:   version,
	}

	return v, nil
}

// CompareVersions compares two version strings
// Returns:
// -1 if v1 < v2
// 0 if v1 == v2
// 1 if v1 > v2
func CompareVersions(v1, v2 string) (int, error) {
	version1, err := ParseVersion(v1)
	if err != nil {
		return 0, err
	}

	version2, err := ParseVersion(v2)
	if err != nil {
		return 0, err
	}

	// Compare major
	if version1.Major != version2.Major {
		if version1.Major < version2.Major {
			return -1, nil
		}
		return 1, nil
	}

	// Compare minor
	if version1.Minor != version2.Minor {
		if version1.Minor < version2.Minor {
			return -1, nil
		}
		return 1, nil
	}

	// Compare patch
	if version1.Patch != version2.Patch {
		if version1.Patch < version2.Patch {
			return -1, nil
		}
		return 1, nil
	}

	// Compare build
	if version1.Build != version2.Build {
		if version1.Build < version2.Build {
			return -1, nil
		}
		return 1, nil
	}

	return 0, nil
}

// Version represents a parsed JDK version
type Version struct {
	Major int
	Minor int
	Patch int
	Build int
	Raw   string
}

// IsLTS checks if this version is an LTS release
// JDK 8, 11, 17, 21, 27, etc. are LTS releases (every 3 years starting from 11)
func (v *Version) IsLTS() bool {
	ltsVersions := []int{8, 11, 17, 21, 27, 33, 39}
	for _, LTS := range ltsVersions {
		if v.Major == LTS {
			return true
		}
	}
	return false
}

// IsGreaterThan checks if this version is greater than the given version
func (v *Version) IsGreaterThan(other *Version) bool {
	result, _ := CompareVersions(v.Raw, other.Raw)
	return result > 0
}

// IsLessThan checks if this version is less than the given version
func (v *Version) IsLessThan(other *Version) bool {
	result, _ := CompareVersions(v.Raw, other.Raw)
	return result < 0
}

// Equals checks if this version equals the given version
func (v *Version) Equals(other *Version) bool {
	result, _ := CompareVersions(v.Raw, other.Raw)
	return result == 0
}

// parseInt converts string to int, returns 0 if empty
func parseInt(s string) int {
	if s == "" {
		return 0
	}
	i, _ := strconv.Atoi(s)
	return i
}

// parseIntDefault converts string to int with default value
func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	i, _ := strconv.Atoi(s)
	return i
}

// ParseVersionError represents an error when parsing a version
type ParseVersionError struct {
	Version string
	Message string
}

func (e *ParseVersionError) Error() string {
	return e.Message + ": " + e.Version
}
