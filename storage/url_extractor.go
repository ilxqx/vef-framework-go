package storage

import (
	"regexp"
	"strings"

	"github.com/dlclark/regexp2"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/set"
)

// Uses regexp2 instead of standard library because backreferences (\1, \2) are needed
// to ensure opening and closing quotes match (e.g., reject src="url').
var (
	htmlImgSrc     = regexp2.MustCompile(`(?i)<img[^>]+src\s*=\s*(["'])([^"']+)\1`, regexp2.None)
	htmlAHref      = regexp2.MustCompile(`(?i)<a[^>]+href\s*=\s*(["'])([^"']+)\1`, regexp2.None)
	htmlVideoSrc   = regexp2.MustCompile(`(?i)<video[^>]+src\s*=\s*(["'])([^"']+)\1`, regexp2.None)
	htmlAudioSrc   = regexp2.MustCompile(`(?i)<audio[^>]+src\s*=\s*(["'])([^"']+)\1`, regexp2.None)
	htmlSourceSrc  = regexp2.MustCompile(`(?i)<source[^>]+src\s*=\s*(["'])([^"']+)\1`, regexp2.None)
	htmlEmbedSrc   = regexp2.MustCompile(`(?i)<embed[^>]+src\s*=\s*(["'])([^"']+)\1`, regexp2.None)
	htmlObjectData = regexp2.MustCompile(`(?i)<object[^>]+data\s*=\s*(["'])([^"']+)\1`, regexp2.None)

	htmlUrlPatterns = []*regexp2.Regexp{
		htmlImgSrc,
		htmlAHref,
		htmlVideoSrc,
		htmlAudioSrc,
		htmlSourceSrc,
		htmlEmbedSrc,
		htmlObjectData,
	}

	// Group 1: attribute name, Group 2: quote type, Group 3: URL value.
	htmlAttrReplacePattern = regexp2.MustCompile(`(?i)(src|href|data)\s*=\s*(["'])([^"']+)\2`, regexp2.None)
)

var (
	markdownImagePattern = regexp.MustCompile(`!\[([^]]*)]\(([^)]+)\)`) // ![alt](url)
	markdownLinkPattern  = regexp.MustCompile(`\[([^]]*)]\(([^)]+)\)`)  // [text](url), allows empty text

	markdownUrlPatterns = []*regexp.Regexp{
		markdownImagePattern,
		markdownLinkPattern,
	}
)

// isRelativeUrl checks if a URL is a relative path (not http:// or https://)
func isRelativeUrl(url string) bool {
	url = strings.TrimSpace(url)

	return url != constants.Empty &&
		!strings.HasPrefix(url, "http://") &&
		!strings.HasPrefix(url, "https://")
}

// extractHtmlUrls extracts all relative URLs from HTML content.
func extractHtmlUrls(content string) []string {
	if content == constants.Empty {
		return nil
	}

	urlSet := set.NewHashSet[string]()

	for _, pattern := range htmlUrlPatterns {
		// regexp2 requires iterative FindNextMatch instead of FindAllStringSubmatch
		match, err := pattern.FindStringMatch(content)
		for match != nil && err == nil {
			// Group 0: entire match, Group 1: quote, Group 2: URL
			groups := match.Groups()
			if len(groups) > 2 {
				url := strings.TrimSpace(groups[2].String())
				if isRelativeUrl(url) {
					urlSet.Add(url)
				}
			}

			match, err = pattern.FindNextMatch(match)
		}
	}

	return urlSet.Values()
}

// replaceHtmlUrls replaces URLs in HTML content based on the replacement map.
func replaceHtmlUrls(content string, replacements map[string]string) string {
	if content == constants.Empty || len(replacements) == 0 {
		return content
	}

	var (
		result    strings.Builder
		lastIndex int
	)

	match, err := htmlAttrReplacePattern.FindStringMatch(content)
	for match != nil && err == nil {
		if groups := match.Groups(); len(groups) > 3 {
			// Group 0: entire match, Group 1: attribute name, Group 2: quote, Group 3: URL
			attrName := groups[1].String()
			quote := groups[2].String()
			oldUrl := groups[3].String()

			_, _ = result.WriteString(content[lastIndex:groups[0].Index])

			if newURL, ok := replacements[oldUrl]; ok {
				// Preserve original quote type to maintain HTML consistency
				_, _ = result.WriteString(attrName)
				_ = result.WriteByte(constants.ByteEquals)
				_, _ = result.WriteString(quote)
				_, _ = result.WriteString(newURL)
				_, _ = result.WriteString(quote)
			} else {
				_, _ = result.WriteString(groups[0].String())
			}

			lastIndex = groups[0].Index + groups[0].Length
		}

		match, err = htmlAttrReplacePattern.FindNextMatch(match)
	}

	_, _ = result.WriteString(content[lastIndex:])

	return result.String()
}

// extractMarkdownUrls extracts all relative URLs from Markdown content.
func extractMarkdownUrls(content string) []string {
	if content == constants.Empty {
		return nil
	}

	urlSet := set.NewHashSet[string]()

	for _, pattern := range markdownUrlPatterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 2 {
				url := strings.TrimSpace(match[2])
				// Markdown allows optional titles: (url "title") or (url 'title')
				// Strip the title to get just the URL
				if idx := strings.IndexAny(url, `"'`); idx > 0 {
					url = strings.TrimSpace(url[:idx])
				}

				if isRelativeUrl(url) {
					urlSet.Add(url)
				}
			}
		}
	}

	return urlSet.Values()
}

// replaceMarkdownUrls replaces URLs in Markdown content based on the replacement map.
func replaceMarkdownUrls(content string, replacements map[string]string) string {
	if content == constants.Empty || len(replacements) == 0 {
		return content
	}

	result := content

	result = markdownImagePattern.ReplaceAllStringFunc(result, func(match string) string {
		if subMatches := markdownImagePattern.FindStringSubmatch(match); len(subMatches) > 2 {
			alt := subMatches[1]
			url := strings.TrimSpace(subMatches[2])

			// Preserve optional title if present
			title := constants.Empty
			if idx := strings.IndexAny(url, `"'`); idx > 0 {
				title = url[idx:]
				url = strings.TrimSpace(url[:idx])
			}

			if newUrl, ok := replacements[url]; ok {
				var sb strings.Builder

				_ = sb.WriteByte(constants.ByteExclamationMark)
				_ = sb.WriteByte(constants.ByteLeftBracket)
				_, _ = sb.WriteString(alt)
				_ = sb.WriteByte(constants.ByteRightBracket)
				_ = sb.WriteByte(constants.ByteLeftParenthesis)
				_, _ = sb.WriteString(newUrl)

				if title != constants.Empty {
					_ = sb.WriteByte(constants.ByteSpace)
					_, _ = sb.WriteString(title)
				}

				_ = sb.WriteByte(constants.ByteRightParenthesis)

				return sb.String()
			}
		}

		return match
	})

	result = markdownLinkPattern.ReplaceAllStringFunc(result, func(match string) string {
		if subMatches := markdownLinkPattern.FindStringSubmatch(match); len(subMatches) > 2 {
			text := subMatches[1]
			url := strings.TrimSpace(subMatches[2])

			// Preserve optional title if present
			title := constants.Empty
			if idx := strings.IndexAny(url, `"'`); idx > 0 {
				title = url[idx:]
				url = strings.TrimSpace(url[:idx])
			}

			if newUrl, ok := replacements[url]; ok {
				var sb strings.Builder

				_ = sb.WriteByte(constants.ByteLeftBracket)
				_, _ = sb.WriteString(text)
				_ = sb.WriteByte(constants.ByteRightBracket)
				_ = sb.WriteByte(constants.ByteLeftParenthesis)
				_, _ = sb.WriteString(newUrl)

				if title != constants.Empty {
					_ = sb.WriteByte(constants.ByteSpace)
					_, _ = sb.WriteString(title)
				}

				_ = sb.WriteByte(constants.ByteRightParenthesis)

				return sb.String()
			}
		}

		return match
	})

	return result
}
