// Package strlist tidies string lists read in as configuration values.
//
// It works on lists that are already split; splitting a comma-separated string is out of
// scope, because that is the configuration library's job (caarlos0/env and the like).
package strlist

import "strings"

// Normalize tidies inconsistencies in a string list read in as a configuration value.
// It trims surrounding whitespace from each element, drops empty elements and duplicates,
// and keeps the order of first appearance.
//
// The splitter only calls `strings.Split` and drops neither whitespace nor empty elements,
// so `"a, b,,a"` arrives as `["a", " b", "", "a"]`. Passed downstream, that makes allowlist
// comparisons miss on a padded element and menus list the same item twice.
//
// It never fills in a default: empty input returns an empty slice. Whether empty is
// acceptable is for the caller's validation to decide, and substituting a default here
// would hide a missing setting.
func Normalize(values []string) []string {
	return normalize(values, nil)
}

// NormalizeFold does what Normalize does, and also lowercases each element and treats
// case differences as duplicates. `["Example.com", "EXAMPLE.com"]` becomes `["example.com"]`.
//
// Use it for settings whose values are case-insensitive, such as hostnames, mail domains
// and allowlists of identifiers. The comparing side must lowercase too before comparing.
// For values where case carries meaning (tokens, API keys and the like), use Normalize.
func NormalizeFold(values []string) []string {
	return normalize(values, strings.ToLower)
}

// normalize is the shared implementation of Normalize and NormalizeFold.
// If fold is not nil, it is applied to each element after trimming.
func normalize(values []string, fold func(string) string) []string {
	seen := make(map[string]struct{}, len(values))
	normalized := make([]string, 0, len(values))

	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if fold != nil {
			v = fold(v)
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		normalized = append(normalized, v)
	}

	return normalized
}
