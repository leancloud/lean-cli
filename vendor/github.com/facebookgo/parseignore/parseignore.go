// Package parseignore implements a subset of the gitignore specification.
//
// This implementation does not support the ** special syntax.
//
// More details for gitignore are available at:
// http://git-scm.com/docs/gitignore
package parseignore

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

var (
	unescape  = regexp.MustCompile(`\\([^\\])`) // Remove the \ from string
	charGroup = regexp.MustCompile(`\[[^\]]+\]`)
)

// Decision is used in tri-state logic
// Look at the constants defined below for more information
type Decision int

const (
	// Pass indicates the Matcher does not have any specific decision about
	// the given path.
	Pass Decision = iota

	// Include indicates the Matcher decided to explicitly include the path.
	// This is used for negation, or patterns that explicitly want to include
	// files when they were otherwise excluded by an earlier rule.
	Include

	// Exclude indicates a pattern matches file path and wants to explicitly
	// exclude it
	Exclude
)

// Matcher matches paths and returns a bool indicating if the path should be
// ignored or not.
type Matcher interface {
	Match(path string, fi os.FileInfo) (Decision, error)
}

type multiMatcher []Matcher

func (m multiMatcher) Match(path string, fi os.FileInfo) (Decision, error) {
	final := Pass
	for _, e := range m {
		decision, err := e.Match(path, fi)
		if err != nil {
			return Pass, err
		}
		if decision != Pass {
			final = decision
		}
	}
	return final, nil
}

// MultiMatcher returns a single Matcher that runs thru the ordered list of
// Matchers, where the last decision wins.
func MultiMatcher(m ...Matcher) Matcher {
	return multiMatcher(m)
}

type inverseMatcher struct {
	m Matcher
}

func (m inverseMatcher) Match(path string, fi os.FileInfo) (Decision, error) {
	decision, err := m.m.Match(path, fi)
	if err != nil {
		return Pass, err
	}
	switch decision {
	case Include:
		return Exclude, nil
	case Exclude:
		return Include, nil
	}
	return Pass, nil
}

// InverseMatcher returns a Matcher that inverts the decision of the given
// matcher.
func InverseMatcher(m Matcher) Matcher {
	return inverseMatcher{m: m}
}

type componentNameMatcher struct {
	Name    string
	DirOnly bool
}

func (m *componentNameMatcher) Match(path string, fi os.FileInfo) (Decision, error) {
	parts := strings.Split(path, "/")
	for i, split := range parts {
		if m.DirOnly && i == len(parts)-1 && !fi.IsDir() {
			return Pass, nil
		}
		if split == m.Name {
			return Exclude, nil
		}
	}
	return Pass, nil
}

// ComponentNameMatcher returns a Matcher that checks if any of the path
// components match the given name.
func ComponentNameMatcher(name string, dirOnly bool) (Matcher, error) {
	if !isFilename(name) {
		return nil, fmt.Errorf("parseignore: %q is an invalid component name", name)
	}
	return &componentNameMatcher{
		Name:    unescape.ReplaceAllString(name, "$1"),
		DirOnly: dirOnly,
	}, nil
}

type fileGlobMatcher struct {
	Glob    string
	DirOnly bool
}

func (m *fileGlobMatcher) Match(path string, fi os.FileInfo) (Decision, error) {
	matched, err := filepath.Match(m.Glob, path)
	if err != nil {
		return Pass, err
	}
	if !matched {
		return Pass, nil
	}
	if m.DirOnly && !fi.IsDir() {
		return Pass, nil
	}
	return Exclude, nil
}

type regexpMatcher struct {
	Regexp  *regexp.Regexp
	DirOnly bool
}

func (m *regexpMatcher) Match(path string, fi os.FileInfo) (Decision, error) {
	if m.Regexp.MatchString(path) {
		if m.DirOnly && !fi.IsDir() {
			return Pass, nil
		}
		return Exclude, nil
	}
	return Pass, nil
}

// GlobMatcher returns a Matcher that checks the path against the provided glob.
func GlobMatcher(glob string, dirOnly bool) (Matcher, error) {
	if strings.Contains(glob, "/") {
		return &fileGlobMatcher{
			Glob:    glob,
			DirOnly: dirOnly,
		}, nil

	}
	re, err := regexp.Compile(translate(glob) + "$")
	if err != nil {
		return nil, fmt.Errorf("parseignore: invalid pattern %q: %s", glob, err)
	}
	return &regexpMatcher{
		Regexp:  re,
		DirOnly: dirOnly,
	}, nil
}

func compilePattern(pattern string) (Matcher, error) {
	originalPattern := pattern
	if strings.Contains(pattern, "**") {
		return nil, fmt.Errorf("parseignore: pattern %q uses the unsupported ** syntax", pattern)
	}

	// trim trailing whitespace unless it has been escaped
	if right := strings.TrimRightFunc(pattern, unicode.IsSpace); !strings.HasSuffix(right, "\\") {
		pattern = right
	}

	// strip ! but record if it's an inverse
	var inverse bool
	if strings.HasPrefix(pattern, "!") {
		inverse = true
		pattern = pattern[1:]
	}

	// trailing slashes indicate patterns should only match a directory
	var dirOnly bool
	if strings.HasSuffix(pattern, "/") {
		dirOnly = true
		pattern = pattern[:len(pattern)-1]
	}

	if pattern == "" {
		return nil, fmt.Errorf("parseignore: pattern %q is invalid", originalPattern)
	}

	var matcher Matcher
	var err error
	if isFilename(pattern) {
		matcher, err = ComponentNameMatcher(pattern, dirOnly)
	} else {
		matcher, err = GlobMatcher(pattern, dirOnly)
	}
	if err != nil {
		return nil, err
	}

	if inverse {
		matcher = InverseMatcher(matcher)
	}
	return matcher, nil
}

// CompilePatterns compiles the patterns contained in the provided contents
// (separated by newlines). It follows a subset of the gitignore specification,
// including rules such as ignoring blank or whitespace only lines. It also
// ignores line that begin with a `#`. Additionally even when we encounter
// errors in compiling patterns we return a Matcher that respects the patterns
// that were successfully compiled.
func CompilePatterns(contents []byte) (Matcher, []error) {
	var matchers []Matcher
	var errors []error
	scanner := bufio.NewScanner(bytes.NewReader(contents))
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		if m, err := compilePattern(line); err != nil {
			errors = append(errors, fmt.Errorf("%s on line %d", err, lineNum))
		} else {
			matchers = append(matchers, m)
		}
	}
	return MultiMatcher(matchers...), errors
}

func isFilename(pattern string) bool {
	var escaped bool
	for _, char := range pattern {
		if char == '/' {
			return false
		}

		if escaped {
			escaped = false
			continue
		}

		switch char {
		case '\\':
			escaped = true
		case '*', '?', '[':
			return false
		}
	}
	return true
}

func translate(pattern string) string {
	var buffer bytes.Buffer

	matches := charGroup.FindAllStringIndex(pattern, -1)
	locations := []int{0}

	for _, match := range matches {
		locations = append(locations, match[0], match[1])
	}
	locations = append(locations, len(pattern))

	for i := 0; i < len(locations)-1; i++ {
		left := locations[i]
		right := locations[i+1]
		if left == right {
			continue
		}
		if i&1 == 1 { // it is a char group
			chunk := pattern[left:right]
			switch string(chunk[1]) {
			case "!":
				chunk = "[^" + chunk[2:]
			case "^":
				chunk = "[\\^" + chunk[2:]
			}
			buffer.WriteString(chunk)
			continue
		}

		var escaped bool
		for _, char := range pattern[left:right] {
			str := string(char)

			if escaped {
				buffer.WriteString(str)
				escaped = false
				continue
			}

			switch str {
			case `\`:
				escaped = true
				buffer.WriteString(str)
			case `*`:
				buffer.WriteString(".*")
			case `?`:
				buffer.WriteString(".")
			default:
				buffer.WriteString(regexp.QuoteMeta(str))
			}
		}
	}
	return buffer.String()
}
