package console

import (
	"regexp"
)

type promptMatcher interface {
	match(line string) bool
	getMatched() interface{}
}

type promptMatcherRegexp struct {
	re      *regexp.Regexp
	matched interface{}
}

func (m *promptMatcherRegexp) match(line string) bool {
	m.matched = nil
	if arr := m.re.FindStringSubmatch(line); arr != nil {
		m.matched = arr
		return true
	}

	return false
}

func (m *promptMatcherRegexp) getMatched() interface{} {
	return m.matched
}

func newPromptRegexpMatcher(pattern string) promptMatcher {
	return &promptMatcherRegexp{
		re: regexp.MustCompile(pattern),
	}
}
