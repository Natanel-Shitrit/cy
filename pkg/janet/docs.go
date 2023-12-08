package janet

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// Docstrings provides docstrings for module methods. This is a mapping from
// the method's Go name to its Janet docstring.
type Docstrings interface {
	Docstrings() map[string]string
}

// Documented provides documentation in the form of a Markdown-formatted
// string. All of the lines following a top-level Markdown header with a title
// that matches a Go method name will be included as the Janet documentation
// for that method.
//
// For example, providing a string that looks like this:
// ```
// # SomeMethod
// This is some documentation
// # SomeMethodB
// This is some other documentation
// ```
// Will result in SomeMethod and SomeMethodB having those docstrings.
type Documented interface {
	Documentation() string
}

// Why oh why did I ever think this was a good idea?
func parseDocstrings(markdown string) (result map[string]string) {
	result = make(map[string]string)

	source := []byte(markdown)
	document := goldmark.
		DefaultParser().
		Parse(text.NewReader(source))

	if document.Type() != ast.TypeDocument || !document.HasChildren() {
		return
	}

	// Make a more sensible array
	children := make([]ast.Node, 0)
	child := document.FirstChild()
	for i := 0; i < document.ChildCount(); i++ {
		children = append(children, child)
		child = child.NextSibling()
	}

	method := ""
	docstring := ""
	for _, child := range children {
		if child.Type() != ast.TypeBlock {
			continue
		}

		if heading, ok := child.(*ast.Heading); ok && heading.Level == 1 {
			if len(method) > 0 {
				result[method] = docstring
				method = ""
				docstring = ""
			}

			method = string(heading.Text(source))
			continue
		}

		if child.HasBlankPreviousLines() {
			docstring += "\n"
		}

		lines := child.Lines()
		for i := 0; i < lines.Len(); i++ {
			segment := lines.At(i)
			docstring += string((&segment).Value(source)) + "\n"
		}
	}

	if len(method) > 0 {
		result[method] = docstring
	}

	return
}
