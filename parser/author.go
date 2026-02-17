package parser

import "strings"

// Author represents a book author with name components
type Author struct {
	FirstName  string
	LastName   string
	MiddleName string
}

// FullName returns the complete author name
func (a Author) FullName() string {
	parts := []string{}
	if a.FirstName != "" {
		parts = append(parts, a.FirstName)
	}
	if a.MiddleName != "" {
		parts = append(parts, a.MiddleName)
	}
	if a.LastName != "" {
		parts = append(parts, a.LastName)
	}
	return strings.Join(parts, " ")
}

// IsEmpty returns true if the author has no name components
func (a Author) IsEmpty() bool {
	return a.FirstName == "" && a.LastName == "" && a.MiddleName == ""
}
