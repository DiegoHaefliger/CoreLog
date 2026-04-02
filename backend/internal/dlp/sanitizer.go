package dlp

import (
	"regexp"
)

// Pre-compiled regexes for maximum performance
var (
	emailRegex    = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	ccRegex       = regexp.MustCompile(`\b(?:\d[ -]*?){13,16}\b`)
	passwordRegex = regexp.MustCompile(`(?i)(password|senha|secret|pass)\s*[:=]\s*['"]?([^'"\s&]+)['"]?`)
	bearerRegex   = regexp.MustCompile(`(?i)(bearer\s+)(eyJ[a-zA-Z0-9-_]+\.[a-zA-Z0-9-_]+\.[a-zA-Z0-9-_]+)`)
)

// Sanitize passes the input text through all redaction filters
func Sanitize(text string) string {
	res := emailRegex.ReplaceAllString(text, "[EMAIL_REDACTED]")
	res = ccRegex.ReplaceAllString(res, "[CC_REDACTED]")
	
	// Password is a bit tricky, we keep the key but redact the value
	res = passwordRegex.ReplaceAllString(res, "${1}=?[SECRET_REDACTED]")
	
	// Tokens keep the "Bearer " prefix
	res = bearerRegex.ReplaceAllString(res, "${1}[TOKEN_REDACTED]")
	
	return res
}
