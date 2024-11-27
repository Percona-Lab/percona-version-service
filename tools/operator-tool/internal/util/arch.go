package util

import "strings"

func GetArchSuffixes() []string {
	return []string{
		"-arm64",
		"-aarch64",
		"-multi",
		"-amd64",
	}
}

func HasArchSuffix(s string) bool {
	for _, suffix := range GetArchSuffixes() {
		if strings.HasSuffix(s, suffix) {
			return true
		}
	}
	return false
}
