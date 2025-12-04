package content

import (
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Article is a minimal parsed article structure for people_military pages.
type Article struct {
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	PublishTime time.Time `json:"publish_time"`
}

// ParsePeopleMilitary parses a HTML page from 人民网-军事，提取标题、正文和发布时间。
// 这里只是雏形实现，后续可以根据真实页面结构调整选择器。
func ParsePeopleMilitary(html string) (*Article, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	var a Article

	// 标题：尝试 h1 或网页 title
	if h1 := doc.Find("h1").First(); h1 != nil && strings.TrimSpace(h1.Text()) != "" {
		a.Title = strings.TrimSpace(h1.Text())
	} else {
		a.Title = strings.TrimSpace(doc.Find("title").First().Text())
	}

	// 正文：根据常见结构尝试几个候选容器
	candidates := []string{
		"#rwb_zw",     // 人民网常见正文 id
		".rm_txt_con", // 另一种正文 class
		".box_con",    // 备用
		".article",    // 通用文章容器
		"body",        // 兜底
	}
	for _, sel := range candidates {
		selection := doc.Find(sel)
		if selection.Length() == 0 {
			continue
		}
		// 优先拼接段落文本，避免把页面上所有导航/脚注一起抓进来
		var paragraphs []string
		selection.Find("p").Each(func(_ int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if text != "" {
				paragraphs = append(paragraphs, text)
			}
		})
		if len(paragraphs) == 0 {
			// 退化为容器整体文本
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

	// 发布时间：尝试常见 class/id
	var publishStr string
	for _, sel := range []string{".rm_txt_time", ".souce span", "#rwb_zw span", ".time", ".pub_time"} {
		publishStr = strings.TrimSpace(doc.Find(sel).First().Text())
		if publishStr != "" {
			break
		}
	}
	if publishStr != "" {
		// 简单尝试几种时间格式，失败就保持零值
		layouts := []string{
			"2006年01月02日 15:04",
			"2006-01-02 15:04:05",
			"2006-01-02",
		}
		for _, layout := range layouts {
			if t, err := time.ParseInLocation(layout, publishStr, time.Local); err == nil {
				a.PublishTime = t
				break
			}
		}
	}

	return &a, nil
}
