package helper

import "strings"

// FoldPortugueseAccent maps accented Latin letters onto basic Latin for case-insensitive matching.
func FoldPortugueseAccent(r rune) rune {
	switch r {
	case 'á', 'à', 'â', 'ã', 'ä':
		return 'a'
	case 'Á', 'À', 'Â', 'Ã', 'Ä':
		return 'A'
	case 'é', 'è', 'ê', 'ë':
		return 'e'
	case 'É', 'È', 'Ê', 'Ë':
		return 'E'
	case 'í', 'ì', 'î', 'ï':
		return 'i'
	case 'Í', 'Ì', 'Î', 'Ï':
		return 'I'
	case 'ó', 'ò', 'ô', 'õ', 'ö':
		return 'o'
	case 'Ó', 'Ò', 'Ô', 'Õ', 'Ö':
		return 'O'
	case 'ú', 'ù', 'û', 'ü':
		return 'u'
	case 'Ú', 'Ù', 'Û', 'Ü':
		return 'U'
	case 'ç':
		return 'c'
	case 'Ç':
		return 'C'
	}
	return r
}

// normalizeHeader lowercases, folds Portuguese accents and strips non-alphanumeric characters.
func normalizeHeader(s string) string {
	var b strings.Builder
	for _, r := range s {
		r = FoldPortugueseAccent(r)
		if r >= 'A' && r <= 'Z' {
			r = r + ('a' - 'A')
		}
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}
