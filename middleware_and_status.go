package core

import "fmt"

type middlewareEntry struct {
	name    string
	wrap    Middleware
	enabled bool
}

// Anonymous functions always append. Functions cannot be safely identified by
// code pointer: closures with different captured values can share that pointer.
func addMiddlewares(entries []middlewareEntry, values []any) []middlewareEntry {
	for _, value := range values {
		entry := middlewareEntry{enabled: true}
		switch m := value.(type) {
		case Middleware:
			entry.wrap = m
		case NamedMiddleware:
			if m.Name == "" {
				panic("core: empty middleware name")
			}
			entry.name, entry.wrap = m.Name, m.Middleware
		default:
			panic(fmt.Sprintf("core: expected middleware function or NamedMiddleware, got %T", value))
		}
		if entry.wrap == nil {
			panic("core: nil middleware")
		}
		replaced := false
		if entry.name != "" {
			for i := range entries {
				if entries[i].name == entry.name {
					entries[i] = entry
					replaced = true
					break
				}
			}
		}
		if !replaced {
			entries = append(entries, entry)
		}
	}
	return entries
}

func excludeMiddlewares(entries []middlewareEntry, names []string) []middlewareEntry {
	for _, name := range names {
		if name == "" {
			panic("core: empty middleware exclusion name")
		}
		found := false
		for i := range entries {
			if entries[i].name == name {
				entries[i].enabled = false
				found = true
				break
			}
		}
		if !found {
			entries = append(entries, middlewareEntry{name: name})
		}
	}
	return entries
}
