package parseignore

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/facebookgo/ensure"
	"github.com/facebookgo/testname"
)

type mockFi struct {
	name  string
	isDir bool
}

func (f mockFi) Name() string {
	return f.name
}

func (f mockFi) Size() int64 {
	panic("Not implemented")
}

func (f mockFi) Mode() os.FileMode {
	if f.isDir {
		return os.ModeDir
	}
	return os.FileMode(0)
}

func (f mockFi) ModTime() time.Time {
	panic("Not implemented")
}

func (f mockFi) IsDir() bool {
	return f.isDir
}

func (f mockFi) Sys() interface{} {
	panic("Not implemented")
}

func makeEmptyRoot(t *testing.T) string {
	prefix := fmt.Sprintf("%s-", testname.Get("parse-cli-"))
	root, err := ioutil.TempDir("", prefix)
	ensure.Nil(t, err)
	return root
}

func TestTranslate(t *testing.T) {
	t.Parallel()

	ensure.DeepEqual(t, translate("*"), ".*")
	ensure.DeepEqual(t, translate("\\*"), "\\*")
	ensure.DeepEqual(t, translate("[*]"), "[*]")

	ensure.DeepEqual(t, translate("?"), ".")
	ensure.DeepEqual(t, translate("\\?"), "\\?")
	ensure.DeepEqual(t, translate("[?]"), "[?]")

	ensure.DeepEqual(t, translate("["), "\\[")
	ensure.DeepEqual(t, translate("\\["), "\\[")
	ensure.DeepEqual(t, translate("]"), "\\]")
	ensure.DeepEqual(t, translate("\\]"), "\\]")
	ensure.DeepEqual(t, translate("[]"), "\\[\\]")
	ensure.DeepEqual(t, translate("[\\]"), "[\\]")
	ensure.DeepEqual(t, translate("[[[]]["), "[[[]\\]\\[")
	ensure.DeepEqual(t, translate("[[[\\]"), "[[[\\]")
	ensure.DeepEqual(t, translate("[a]"), "[a]")
	ensure.DeepEqual(t, translate("[a-b]"), "[a-b]")
	ensure.DeepEqual(t, translate("[!a-b]"), "[^a-b]")
	ensure.DeepEqual(t, translate("[^a-b]"), "[\\^a-b]")
	ensure.DeepEqual(t, translate("[a^b-]"), "[a^b-]")

	ensure.DeepEqual(t, translate("*.txt"), ".*\\.txt")
}

func TestIsFileName(t *testing.T) {
	t.Parallel()

	ensure.True(t, isFilename("file"))
	ensure.True(t, isFilename("star\\*"))
	ensure.True(t, isFilename("bracket\\["))
	ensure.True(t, isFilename("question\\?"))

	ensure.False(t, isFilename("*"))
	ensure.False(t, isFilename("?"))
	ensure.False(t, isFilename("["))
}

func TestEmptyPatterns(t *testing.T) {
	t.Parallel()

	ignores := " \t\n\r\n\n"

	_, errs := CompilePatterns([]byte(ignores))
	ensure.DeepEqual(t, len(errs), 0)

	ignores = `
# This is a comment
`
	_, errs = CompilePatterns([]byte(ignores))
	ensure.DeepEqual(t, len(errs), 0)

	ignores = `
**/pattern
pattern/**/pattern
pattern/**/a/**/pattern
pattern/**
`

	_, errs = CompilePatterns([]byte(ignores))
	ensure.DeepEqual(t, len(errs), 4)
	for _, err := range errs {
		ensure.Err(t, err, regexp.MustCompile(`uses the unsupported \*\* syntax`))
	}

	ignores = `
!
`

	_, errs = CompilePatterns([]byte(ignores))
	ensure.DeepEqual(t, len(errs), 1)
	ensure.Err(t, errs[0], regexp.MustCompile(`pattern "!" is invalid`))

	ignores = `
/
`

	_, errs = CompilePatterns([]byte(ignores))
	ensure.DeepEqual(t, len(errs), 1)
	ensure.Err(t, errs[0], regexp.MustCompile(`pattern "/" is invalid`))
}

func TestFileNamePatterns(t *testing.T) {
	t.Parallel()

	ignores := `	
file
a\*\ \ .txt
b\?.html
c}.ll
!selected/file
`

	testCases := []struct {
		path    string
		exclude bool
	}{
		{"file", true},
		{"a*  .txt", true},
		{"b?.html", true},
		{"c}.ll", true},
		{"d.txt", false},
		{"selected/file", false},
		{"a\\*\\ \\ .txt", false},
	}

	matcher, errs := CompilePatterns([]byte(ignores))
	ensure.DeepEqual(t, len(errs), 0)

	for _, testCase := range testCases {
		state, err := matcher.Match(testCase.path, mockFi{})
		ensure.Nil(t, err)
		ensure.DeepEqual(t, state == Exclude, testCase.exclude)
		state, err = matcher.Match(testCase.path, mockFi{isDir: true})
		ensure.Nil(t, err)
		ensure.DeepEqual(t, state == Exclude, testCase.exclude)
	}

	ignores =
		`	
file/
a\*\ \ .txt/
b\?.html/
c}.ll/
!selected/file/
`

	matcher, errs = CompilePatterns([]byte(ignores))
	ensure.DeepEqual(t, len(errs), 0)

	for _, testCase := range testCases {
		state, err := matcher.Match(testCase.path, mockFi{})
		ensure.Nil(t, err)
		ensure.DeepEqual(t, state, Pass)
		state, err = matcher.Match(testCase.path, mockFi{isDir: true})
		ensure.Nil(t, err)
		ensure.DeepEqual(t, state == Exclude, testCase.exclude)
	}
}

func TestFileGlobPatterns(t *testing.T) {
	t.Parallel()
	ignores := `	
a/b
a/*/b
a/b*.txt
!a/bc.txt
`

	matcher, errs := CompilePatterns([]byte(ignores))
	ensure.DeepEqual(t, len(errs), 0)

	testCases := []struct {
		path     string
		excluded bool
	}{
		{"a/b", true},
		{"a/c/b", true},
		{"a/b.txt", true},
		{"a/bd.txt", true},
		{"a/bc.txt", false},
		{"a/b/c.txt", false},
		{"a/b/c/d", false},
	}

	for _, testCase := range testCases {
		state, err := matcher.Match(testCase.path, mockFi{})
		ensure.Nil(t, err)
		ensure.DeepEqual(t, state == Exclude, testCase.excluded)
		state, err = matcher.Match(testCase.path, mockFi{isDir: true})
		ensure.Nil(t, err)
		ensure.DeepEqual(t, state == Exclude, testCase.excluded)
	}

	ignores =
		`	
a/b/
a/*/b/
a/b*.txt/
!a/bc.txt/
`

	matcher, errs = CompilePatterns([]byte(ignores))
	ensure.DeepEqual(t, len(errs), 0)

	for _, testCase := range testCases {
		state, err := matcher.Match(testCase.path, mockFi{})
		ensure.Nil(t, err)
		ensure.DeepEqual(t, state, Pass)
		state, err = matcher.Match(testCase.path, mockFi{isDir: true})
		ensure.Nil(t, err)
		ensure.DeepEqual(t, state == Exclude, testCase.excluded)
	}
}

func TestListFiles(t *testing.T) {
	t.Parallel()

	testsInfo := []struct {
		Name  string
		IsDir bool
	}{
		{"tester", true},
		{"tester/escape\\.yo", false},
		{"tester/test", false},
		{"tester/star\\*.txt", false},
		{"tester/hello", false},
		{"tester/hello.txt", false},
		{"tester/hello.txter", false},

		{"tester/inside", true},
		{"tester/inside/test", true},
		{"tester/inside/tester", true},
		{"tester/inside/tester/test", true},

		{"tester/insider", true},
		{"tester/insider/tester", true},
		{"tester/insider/tester/test", false},
	}

	testCases := []struct {
		ignores         string
		expectedMatches []string
	}{
		{
			ignores: `
*.txt
`,
			expectedMatches: []string{
				"tester/hello.txt",
				"tester/star\\*.txt",
			},
		},
		{
			ignores: `
test/
`,
			expectedMatches: []string{
				"tester/inside/test",
				"tester/inside/tester/test",
			},
		},
		{
			ignores: `
test
`,
			expectedMatches: []string{
				"tester/test",
				"tester/inside/test",
				"tester/inside/tester/test",
				"tester/insider/tester/test",
			},
		},
	}

	for _, testCase := range testCases {
		matcher, errs := CompilePatterns([]byte(testCase.ignores))
		ensure.DeepEqual(t, len(errs), 0)

		expectedMatches := make(map[string]struct{})
		for _, expectedMatch := range testCase.expectedMatches {
			expectedMatches[expectedMatch] = struct{}{}
		}
		for _, testInfo := range testsInfo {
			matched, err := matcher.Match(testInfo.Name, mockFi{isDir: testInfo.IsDir})
			ensure.Nil(t, err)

			if _, ok := expectedMatches[testInfo.Name]; ok {
				ensure.DeepEqual(t, matched, Exclude)
			} else {
				ensure.DeepEqual(t, matched, Pass)
			}
		}
	}
}
