package builder

import (
	"encoding/base64"
	"strings"
)

// Base64KeywordEncoding replaces the characters that base64 shares with Go syntax, so a hash can be used in an identifier.
var Base64KeywordEncoding = keywordEncoding{
	PlusSign:  "Σ",
	SlashSign: "Δ",
	EqualSign: "Ξ",
}

// keywordEncoding is the replacement set for base64 output.
type keywordEncoding struct {
	PlusSign  string
	SlashSign string
	EqualSign string
}

func (e keywordEncoding) EncodeToString(bs []byte) string {
	str := base64.StdEncoding.EncodeToString(bs)
	str = strings.ReplaceAll(str, "+", e.PlusSign)
	str = strings.ReplaceAll(str, "/", e.SlashSign)
	return strings.ReplaceAll(str, "=", e.EqualSign)
}
