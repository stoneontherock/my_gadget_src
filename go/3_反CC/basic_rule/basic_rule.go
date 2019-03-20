package basic_rule

import (
	"regexp"
	"strings"
)

type Rule struct {
	Regexp *regexp.Regexp
}

func NewRule(str string, isWildcard bool) (Rule, error) {
	var r Rule

	if isWildcard {
		str = strings.Replace(str, ".", `\.`, -1)
		str = strings.Replace(str, "*", ".*", -1)
		str = strings.Replace(str, "?", ".", -1)
	}

	var err error
	r.Regexp, err = regexp.Compile(str)
	return r, err
}

func (r *Rule) Match(s string) bool {
	//logrus.Debugf("rule.Regex=%v", r.Regexp)
	return r.Regexp.MatchString(s)
}
