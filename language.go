package main

import "strings"

func ParseLanguage(label string) Language {
	switch strings.ToLower(strings.TrimSpace(label)) {
	case "go":
		return LangGo
	case "python":
		return LangPython
	case "rust":
		return LangRust
	default:
		return LangUnknown
	}
}
