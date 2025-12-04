package content

import "fmt"

var ErrUnsupportedSource = fmt.Errorf("unsupported source code")

// Parse is a unified entry point for parsing different news sources by source code.
// It routes to site-specific parsers based on sourceCode.
func Parse(sourceCode, html string) (*Article, error) {
	switch sourceCode {
	case "people_military":
		return ParsePeopleMilitary(html)
	case "xinhua_military":
		return ParseXinhuaMilitary(html)
	case "gmw_military":
		return ParseGmwMilitary(html)
	default:
		return nil, ErrUnsupportedSource
	}
}
