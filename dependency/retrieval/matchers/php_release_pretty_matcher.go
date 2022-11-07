package matchers

import (
	"fmt"

	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
	"github.com/paketo-buildpacks/php-dist/retrieval"
)

func NewPhpReleasePrettyMatcher(pretty retrieval.PhpReleasePretty) types.GomegaMatcher {
	return &PhpReleasePrettyMatcher{
		expected: pretty,
	}
}

type PhpReleasePrettyMatcher struct {
	expected retrieval.PhpReleasePretty
}

func (m *PhpReleasePrettyMatcher) Match(actual interface{}) (bool, error) {
	pretty, ok := actual.(retrieval.PhpReleasePretty)
	if !ok {
		return false, fmt.Errorf("PhpReleasePrettyMatcher expects an retrieval.PhpReleasePretty, received %T", actual)
	}

	semverMatches := pretty.SemverVersion == m.expected.SemverVersion ||
		(pretty.SemverVersion != nil && m.expected.SemverVersion != nil && pretty.SemverVersion.Equal(m.expected.SemverVersion))

	releaseDateMatches := pretty.ReleaseDate == m.expected.ReleaseDate ||
		(pretty.ReleaseDate != nil && m.expected.ReleaseDate != nil && pretty.ReleaseDate.Equal(*m.expected.ReleaseDate))

	return semverMatches &&
		releaseDateMatches &&
		pretty.Sha256 == m.expected.Sha256, nil
}

func (m *PhpReleasePrettyMatcher) FailureMessage(actual interface{}) string {
	return format.Message(actual, "to equal", m.expected)
}

func (m *PhpReleasePrettyMatcher) NegatedFailureMessage(actual interface{}) string {
	return format.Message(actual, "not to equal", m.expected)
}
