package treenews

import (
	"regexp"
	"strings"
)

var (
	reBlacklistStrong = regexp.MustCompile(`(?i)\b(termination|terminate|ended|end\s+of\s+support|delist(?:ing)?|suspend(?:ed|ing|sion)?|halt|pause|maintenance|upgrade|outage|incident|issue|bug|resume|resumption|reopen(?:ing)?|postpone(?:d|ment)?|delay(?:ed|s)?)\b`)
	reParens          = regexp.MustCompile(`\(([^()]*)\)`)
	reTokenSplit      = regexp.MustCompile(`[^A-Za-z0-9._+\-/]+`)
)

var stopwords = map[string]struct{}{
	"BTC":    {},
	"USDT":   {},
	"MARKET": {},
	"KRW":    {},
}

func passKRWFilter(record map[string]any) bool {
	en := strings.ToUpper(toString(record["en"]))
	if en == "" {
		return false
	}
	if reBlacklistStrong.MatchString(en) {
		return false
	}
	return strings.Contains(en, "KRW")
}

func extractSymbols(record map[string]any) []string {
	tokens := append(
		tokensFromParens(toString(record["title"])),
		tokensFromParens(toString(record["en"]))...)
	seen := make(map[string]struct{}, len(tokens))
	out := make([]string, 0, len(tokens))
	for _, t := range tokens {
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}

func tokensFromParens(text string) []string {
	if text == "" {
		return nil
	}
	matches := reParens.FindAllStringSubmatch(text, -1)
	out := make([]string, 0)
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		inner := m[1]
		if strings.Contains(strings.ToUpper(inner), "KRW") {
			continue
		}
		for _, token := range reTokenSplit.Split(inner, -1) {
			tok := strings.ToUpper(strings.TrimSpace(token))
			if tok == "" {
				continue
			}
			if _, ok := stopwords[tok]; ok {
				continue
			}
			out = append(out, tok)
		}
	}
	return out
}

func isUpbitSource(v string) bool {
	if v == "" {
		return false
	}
	return strings.Contains(strings.ToLower(v), "upbit")
}

func isBinanceSource(v string) bool {
	return strings.EqualFold(strings.TrimSpace(v), "Binance EN")
}

func passBinanceSpotFilter(record map[string]any) bool {
	if !isBinanceSource(toString(record["source"])) {
		return false
	}
	title := strings.ToLower(toString(record["title"]))
	en := strings.ToLower(toString(record["en"]))
	switch {
	case strings.Contains(title, "binance will list"), strings.Contains(en, "binance will list"):
		return true
	case strings.Contains(title, "introducing") && strings.Contains(title, "hodler airdrops"):
		return true
	case strings.Contains(en, "introducing") && strings.Contains(en, "hodler airdrops"):
		return true
	default:
		return false
	}
}

func binanceSpotSymbols(event map[string]any) []string {
	if !passBinanceSpotFilter(event) {
		return nil
	}
	syms := extractSymbols(event)
	if len(syms) == 0 {
		return nil
	}
	return syms
}

func routeSymbols(event map[string]any) (string, []string) {
	source := strings.ToLower(strings.TrimSpace(toString(event["source"])))
	if source == "" {
		source = strings.ToLower(strings.TrimSpace(toString(event["type"])))
	}
	switch {
	case strings.Contains(source, "upbit"):
		syms := upbitKRWSymbols(event)
		if len(syms) > 0 {
			return "upbit", syms
		}
	case source == "binance en":
		syms := binanceSpotSymbols(event)
		if len(syms) > 0 {
			return "binance", syms
		}
	}
	return "", nil
}

func upbitKRWSymbols(event map[string]any) []string {
	source := toString(event["source"])
	if source == "" {
		source = toString(event["type"])
	}
	if !isUpbitSource(source) {
		return nil
	}
	if !passKRWFilter(event) {
		return nil
	}
	syms := extractSymbols(event)
	if len(syms) == 0 {
		return nil
	}
	return syms
}
