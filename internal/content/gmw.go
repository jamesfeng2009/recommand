package content

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ParseGmwMilitary parses a Guangming military news page in a best-effort way.
func ParseGmwMilitary(html string) (*Article, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	var a Article

	if h1 := strings.TrimSpace(doc.Find("h1").First().Text()); h1 != "" {
		a.Title = h1
	} else {
		a.Title = strings.TrimSpace(doc.Find("title").First().Text())
	}

	candidates := []string{
		"#contentMain", "#content", ".article", ".wrap", "body",
	}
	for _, sel := range candidates {
		selection := doc.Find(sel)
		if selection.Length() == 0 {
			continue
		}
		var paragraphs []string
		selection.Find("p").Each(func(_ int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if text != "" {
				paragraphs = append(paragraphs, text)
			}
		})
		if len(paragraphs) == 0 {
			text := strings.TrimSpace(selection.Text())
			if text == "" {
				continue
			}
			a.Content = text
		} else {
			a.Content = strings.Join(paragraphs, "\n")
		}
		break
	}

	// leave PublishTime zero for now; can be improved with real page samples

	return &a, nil
}
