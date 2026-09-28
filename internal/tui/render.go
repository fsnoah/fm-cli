package tui

import (
	"fmt"
	"regexp"
	"strings"

	"fm-cli/internal/api"
	"fm-cli/internal/model"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/net/html"
)

// filterEmails returns emails matching the search query
func filterEmails(emails []model.Email, query string) []model.Email {
	if query == "" {
		return emails
	}
	query = strings.ToLower(query)
	var matches []model.Email
	for _, e := range emails {
		// Match against subject, from, to, or preview
		if strings.Contains(strings.ToLower(e.Subject), query) ||
			strings.Contains(strings.ToLower(e.From), query) ||
			strings.Contains(strings.ToLower(e.To), query) ||
			strings.Contains(strings.ToLower(e.Preview), query) {
			matches = append(matches, e)
		}
	}
	return matches
}

// filterEvents returns events matching the search query
func filterEvents(events []model.CalendarEvent, query string) []model.CalendarEvent {
	if query == "" {
		return events
	}
	query = strings.ToLower(query)
	var matches []model.CalendarEvent
	for _, e := range events {
		// Match against title, description, or location
		if strings.Contains(strings.ToLower(e.Title), query) ||
			strings.Contains(strings.ToLower(e.Description), query) ||
			strings.Contains(strings.ToLower(e.Location), query) {
			matches = append(matches, e)
		}
	}
	return matches
}

// filterContactsAll returns all contacts matching the search query (no limit)
func filterContactsAll(contacts []model.Contact, query string) []model.Contact {
	if query == "" {
		return contacts
	}
	query = strings.ToLower(query)
	var matches []model.Contact
	for _, c := range contacts {
		// Match against name, email, company, or phone
		if strings.Contains(strings.ToLower(c.FullName), query) ||
			strings.Contains(strings.ToLower(c.Company), query) ||
			strings.Contains(strings.ToLower(c.Notes), query) {
			matches = append(matches, c)
			continue
		}
		matched := false
		for _, e := range c.Emails {
			if strings.Contains(strings.ToLower(e.Email), query) {
				matches = append(matches, c)
				matched = true
				break
			}
		}
		if matched {
			continue
		}
		for _, p := range c.Phones {
			if strings.Contains(p.Number, query) {
				matches = append(matches, c)
				break
			}
		}
	}
	return matches
}

// filterContacts returns contacts matching the search query (limited for autocomplete)
func filterContacts(contacts []model.Contact, query string) []model.Contact {
	if query == "" {
		return nil
	}
	query = strings.ToLower(query)
	var matches []model.Contact
	for _, c := range contacts {
		// Match against name or email
		if strings.Contains(strings.ToLower(c.FullName), query) {
			matches = append(matches, c)
			continue
		}
		for _, e := range c.Emails {
			if strings.Contains(strings.ToLower(e.Email), query) {
				matches = append(matches, c)
				break
			}
		}
		if len(matches) >= 5 { // Limit suggestions
			break
		}
	}
	return matches
}

// renderBody rebuilds the reader's lines — headers, divider, wrapped body —
// for the open email at the current width. The reader scrolls over these,
// so HTML is converted once per load or resize rather than on every frame.
func (m *Model) renderBody() {
	if m.state != viewBody {
		return
	}
	e := m.openEmail
	width := m.contentWidth()
	label := contactFieldLabelStyle.Width(9)
	field := func(name, value string) string {
		return label.Render(name) + lipgloss.NewStyle().Width(width-9).Render(api.CleanLine(value))
	}

	var lines []string
	lines = append(lines, strings.Split(lipgloss.NewStyle().Bold(true).Foreground(fgColor).Width(width).Render(api.CleanLine(e.Subject)), "\n")...)
	lines = append(lines, strings.Split(field("From", e.From), "\n")...)
	lines = append(lines, field("Date", e.Date))
	if m.showDetails {
		for _, f := range [][2]string{{"To", e.To}, {"Cc", e.Cc}, {"Bcc", e.Bcc}, {"Reply-To", e.ReplyTo}, {"ID", e.ID}} {
			if f[1] != "" {
				lines = append(lines, strings.Split(field(f[0], f[1]), "\n")...)
			}
		}
		lines = append(lines, field("Folders", strings.Join(m.mailboxNames(e.MailboxIDs), ", ")))
	}
	lines = append(lines, emailDateStyle.Render(strings.Repeat("─", width)), "")

	bodyText := renderEmailBody(m.bodyContent, m.htmlBody, width)
	lines = append(lines, strings.Split(wrapTextParagraphs(bodyText, width), "\n")...)
	m.bodyLines = lines
	m.scrollBody(0)
}

// mailboxNames maps folder ids to names, keeping unknown ids as they are.
func (m Model) mailboxNames(ids []string) []string {
	names := make([]string, 0, len(ids))
	for _, id := range ids {
		name := id
		for _, mb := range m.mailboxes {
			if mb.ID == id {
				name = mb.Name
				break
			}
		}
		names = append(names, name)
	}
	return names
}

// Helper to make links clickable (OSC 8)
// shortenURL returns a display-friendly shortened URL
func shortenURL(url string, maxLen int) string {
	if len(url) <= maxLen {
		return url
	}

	// Try to extract domain and show domain + "..."
	// e.g., https://click.redditmail.com/CL0/https%3A... -> click.redditmail.com/...
	re := regexp.MustCompile(`^(https?://)([^/]+)(.*)$`)
	matches := re.FindStringSubmatch(url)
	if matches != nil {
		domain := matches[2]
		path := matches[3]

		// If domain alone is short enough, show domain + truncated path
		if len(domain) < maxLen-4 {
			remaining := maxLen - len(domain) - 4 // 4 for "..." and "/"
			if remaining > 0 && len(path) > 0 {
				if len(path) > remaining {
					return domain + path[:remaining] + "..."
				}
				return domain + path
			}
			return domain + "/..."
		}
		return domain[:maxLen-3] + "..."
	}

	return url[:maxLen-3] + "..."
}

func linkify(text string) string {
	// 1. Convert Markdown links: [Title](URL) -> OSC 8 link
	reMD := regexp.MustCompile(`\[([^\]]+)\]\((https?://[^)]+)\)`)
	text = reMD.ReplaceAllStringFunc(text, func(match string) string {
		parts := reMD.FindStringSubmatch(match)
		if len(parts) < 3 || !api.IsSafeLinkURL(parts[2]) {
			return match
		}
		return "\x1b]8;;" + parts[2] + "\x1b\\" + parts[1] + "\x1b]8;;\x1b\\"
	})

	// 2. Convert Bare URLs: https://google.com -> OSC 8 link with shortened display
	if strings.Contains(text, "[Converted HTML]") {
		return text
	}

	// Plain text mode: Wrap all bare URLs with shortened display text
	reURL := regexp.MustCompile(`(https?://[^\s()<>"]+)`)
	text = reURL.ReplaceAllStringFunc(text, func(url string) string {
		if !api.IsSafeLinkURL(url) {
			return url
		}
		display := shortenURL(url, 50)
		return fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", url, display)
	})
	return text
}

// htmlToText converts HTML to readable plain text using a proper parser
func htmlToText(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return stripHTMLFallback(htmlContent)
	}

	var buf strings.Builder
	var extractText func(*html.Node)

	// Helper to get attribute value
	getAttr := func(n *html.Node, key string) string {
		for _, attr := range n.Attr {
			if attr.Key == key {
				return attr.Val
			}
		}
		return ""
	}

	extractText = func(n *html.Node) {
		// Skip style, script, head elements entirely
		if n.Type == html.ElementNode {
			switch n.Data {
			case "style", "script", "head", "noscript":
				return
			case "br":
				buf.WriteString("\n")
				return
			case "p", "div", "tr", "li", "h1", "h2", "h3", "h4", "h5", "h6", "blockquote":
				buf.WriteString("\n")
			case "td", "th":
				buf.WriteString(" ")
			case "a":
				// Handle links - extract text and href
				href := getAttr(n, "href")
				if href != "" && strings.HasPrefix(href, "http") {
					// Get link text
					var linkText strings.Builder
					var extractLinkText func(*html.Node)
					extractLinkText = func(ln *html.Node) {
						if ln.Type == html.TextNode {
							text := strings.TrimSpace(ln.Data)
							if text != "" {
								linkText.WriteString(text)
							}
						}
						for c := ln.FirstChild; c != nil; c = c.NextSibling {
							extractLinkText(c)
						}
					}
					extractLinkText(n)

					text := strings.TrimSpace(linkText.String())
					if text != "" && text != href && !strings.HasPrefix(text, "http") {
						// Show as "text (shortened_url)"
						buf.WriteString(text)
						buf.WriteString(" (")
						buf.WriteString(shortenURL(href, 40))
						buf.WriteString(") ")
					} else {
						// Just show shortened URL
						buf.WriteString(shortenURL(href, 50))
						buf.WriteString(" ")
					}
					return // Don't process children again
				}
			}
		}

		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			if text != "" {
				buf.WriteString(text)
				buf.WriteString(" ")
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractText(c)
		}

		// Add newline after block elements
		if n.Type == html.ElementNode {
			switch n.Data {
			case "p", "div", "tr", "li", "h1", "h2", "h3", "h4", "h5", "h6", "blockquote":
				buf.WriteString("\n")
			}
		}
	}

	extractText(doc)

	result := buf.String()

	// Clean up whitespace
	reSpaces := regexp.MustCompile(`[ \t]+`)
	result = reSpaces.ReplaceAllString(result, " ")
	reNewlines := regexp.MustCompile(`\n[ \t]+`)
	result = reNewlines.ReplaceAllString(result, "\n")
	reMultiNewlines := regexp.MustCompile(`\n{3,}`)
	result = reMultiNewlines.ReplaceAllString(result, "\n\n")

	// Remove zero-width characters often used in spam
	result = strings.ReplaceAll(result, "\u200b", "") // zero-width space
	result = strings.ReplaceAll(result, "\u200c", "") // zero-width non-joiner
	result = strings.ReplaceAll(result, "\u200d", "") // zero-width joiner
	result = strings.ReplaceAll(result, "\ufeff", "") // BOM

	return strings.TrimSpace(result)
}

// stripHTMLFallback is a simple regex-based fallback
func stripHTMLFallback(htmlContent string) string {
	// Remove style and script blocks entirely
	reStyle := regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	htmlContent = reStyle.ReplaceAllString(htmlContent, "")
	reScript := regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	htmlContent = reScript.ReplaceAllString(htmlContent, "")

	// Replace common block elements with newlines
	reBlock := regexp.MustCompile(`(?i)</(p|div|tr|li|h[1-6])>`)
	htmlContent = reBlock.ReplaceAllString(htmlContent, "\n")
	reBr := regexp.MustCompile(`(?i)<br\s*/?>`)
	htmlContent = reBr.ReplaceAllString(htmlContent, "\n")

	// Remove all remaining tags
	reTags := regexp.MustCompile(`<[^>]+>`)
	htmlContent = reTags.ReplaceAllString(htmlContent, "")

	// Decode common HTML entities
	htmlContent = html.UnescapeString(htmlContent)

	// Collapse whitespace
	reSpaces := regexp.MustCompile(`[ \t]+`)
	htmlContent = reSpaces.ReplaceAllString(htmlContent, " ")
	reNewlines := regexp.MustCompile(`\n{3,}`)
	htmlContent = reNewlines.ReplaceAllString(htmlContent, "\n\n")

	return strings.TrimSpace(htmlContent)
}

// renderEmailBody renders the email body, converting HTML to plain text
func renderEmailBody(textBody, htmlBody string, width int) string {
	// Bodies are cleaned at the API boundary; cached bodies from older
	// versions are cleaned here so a stored escape sequence never renders.
	textBody = api.CleanText(textBody)
	htmlBody = api.CleanText(htmlBody)
	// Check if textBody looks like HTML (contains HTML tags)
	isHTMLBody := strings.Contains(textBody, "<html") ||
		strings.Contains(textBody, "<table") ||
		strings.Contains(textBody, "<div") ||
		strings.Contains(textBody, "<td") ||
		strings.Contains(textBody, "<!DOCTYPE")

	// If we have clean text body (not HTML), prefer it
	if textBody != "" && !strings.HasPrefix(textBody, "[Converted HTML]") && !isHTMLBody {
		return linkify(textBody)
	}

	// For HTML content, convert to plain text
	// Try htmlBody first, then textBody if it's HTML
	contentToConvert := htmlBody
	if contentToConvert == "" && isHTMLBody {
		contentToConvert = textBody
	}

	if contentToConvert != "" {
		text := htmlToText(contentToConvert)
		if text != "" {
			return linkify(text)
		}
	}

	// Fall back to text body with linkify (even if it's converted HTML)
	if textBody != "" {
		return linkify(textBody)
	}

	return "(No content)"
}

// wrapText wraps text to fit within maxWidth characters
func wrapText(text string, maxWidth int) string {
	if len(text) <= maxWidth {
		return text
	}

	var result strings.Builder
	words := strings.Fields(text)
	lineLen := 0

	for i, word := range words {
		wordLen := len(word)
		if lineLen > 0 && lineLen+wordLen+1 > maxWidth {
			result.WriteString("\n")
			lineLen = 0
		}
		if lineLen > 0 {
			result.WriteString(" ")
			lineLen++
		}
		result.WriteString(word)
		lineLen += wordLen

		// Handle very long words
		if wordLen > maxWidth && i < len(words)-1 {
			result.WriteString("\n")
			lineLen = 0
		}
	}

	return result.String()
}

// wrapTextParagraphs wraps text while preserving paragraph structure
func wrapTextParagraphs(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return text
	}

	var result strings.Builder
	paragraphs := strings.Split(text, "\n")

	for i, para := range paragraphs {
		para = strings.TrimSpace(para)

		// Empty lines (paragraph breaks) are preserved
		if para == "" {
			result.WriteString("\n")
			continue
		}

		// Wrap the paragraph
		wrapped := wrapText(para, maxWidth)
		result.WriteString(wrapped)

		// Add newline after paragraph unless it's the last one
		if i < len(paragraphs)-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}
