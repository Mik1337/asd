// Package parse extracts the command and description from the model's reply.
//
// Expected shape:
//
//	[description]: shows the remote named origin and its branches
//	[command]: git remote show origin
//
// Parsing is lenient: tags may appear in any order, the description may span
// several lines (rich mode), and stray code fences are stripped. If no
// [command]: tag is found at all, the first non-empty line is treated as the
// command so the user is never left with nothing.
package parse

import "strings"

const (
	cmdTag  = "[command]:"
	descTag = "[description]:"
)

// Tags returns the command and description found in s.
func Tags(s string) (command, description string) {
	s = stripFences(s)
	lines := strings.Split(s, "\n")

	var desc []string
	inDesc := false
	for _, ln := range lines {
		t := strings.TrimSpace(ln)
		low := strings.ToLower(t)
		switch {
		case strings.HasPrefix(low, cmdTag):
			command = strings.TrimSpace(t[len(cmdTag):])
			inDesc = false
		case strings.HasPrefix(low, descTag):
			desc = append(desc, strings.TrimSpace(t[len(descTag):]))
			inDesc = true
		case inDesc && t != "":
			desc = append(desc, t)
		}
	}
	description = strings.TrimSpace(strings.Join(desc, "\n"))

	if command == "" {
		// Fallback: the model ignored the tags. Treat the first real line as
		// the command and drop the (untrustworthy) description.
		command = firstNonEmpty(lines)
		description = ""
	}
	return command, description
}

func firstNonEmpty(lines []string) string {
	for _, l := range lines {
		if t := strings.TrimSpace(l); t != "" {
			return t
		}
	}
	return ""
}

// stripFences removes a single wrapping ```...``` block if present.
func stripFences(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "```") {
		return s
	}
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[i+1:]
	}
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}
