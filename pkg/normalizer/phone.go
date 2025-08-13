package normalizer

import "unicode"

func NormalizePhone(raw string) (string, string, bool) {
	if raw == "" {
		return "", "", false
	}
	digits := make([]rune, 0, len(raw))
	for _, r := range raw {
		if unicode.IsDigit(r) {
			digits = append(digits, r)
		}
	}
	if len(digits) < 7 || len(digits) > 15 {
		return "", "", false
	}
	ds := string(digits)
	return "+" + ds, ds, true
}
