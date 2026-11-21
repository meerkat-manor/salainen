package securepath

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Resolve(root, key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("no secret path provided")
	}

	name := strings.SplitN(key, "|", 2)[0]
	if name == "" {
		return "", fmt.Errorf("empty secret path")
	}
	if strings.ContainsRune(name, 0) {
		return "", fmt.Errorf("invalid character in secret path")
	}

	normalized := strings.ReplaceAll(name, "\\", "/")
	for _, part := range strings.Split(normalized, "/") {
		if part == ".." {
			return "", fmt.Errorf("paths cannot contain '..' to prevent path traversal")
		}
	}
	if filepath.IsAbs(name) || strings.HasPrefix(normalized, "/") {
		return "", fmt.Errorf("absolute secret paths are not allowed")
	}

	if root == "" && (strings.HasPrefix(name, "~/") || strings.HasPrefix(name, "~\\")) {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		root = homeDir
		name = name[2:]
	}

	base := root
	if base == "" {
		base = "."
	}
	absBase, err := filepath.Abs(base)
	if err != nil {
		return "", err
	}
	absCandidate, err := filepath.Abs(filepath.Join(absBase, name))
	if err != nil {
		return "", err
	}

	rel, err := filepath.Rel(absBase, absCandidate)
	if err != nil {
		return "", err
	}
	if isOutside(rel) {
		return "", fmt.Errorf("secret path escapes the configured root")
	}

	return resolveExisting(absBase, absCandidate)
}

func resolveExisting(base, candidate string) (string, error) {
	baseExisting, _ := existingAncestor(base)
	candidateExisting, candidateSuffix := existingAncestor(candidate)

	resolvedBase, err := filepath.EvalSymlinks(baseExisting)
	if err != nil {
		return "", err
	}
	resolvedCandidate, err := filepath.EvalSymlinks(candidateExisting)
	if err != nil {
		return "", err
	}

	rel, err := filepath.Rel(resolvedBase, resolvedCandidate)
	if err != nil || isOutside(rel) {
		return "", fmt.Errorf("secret path escapes the configured root")
	}

	final := resolvedCandidate
	for _, part := range candidateSuffix {
		final = filepath.Join(final, part)
	}
	return final, nil
}

func existingAncestor(path string) (string, []string) {
	cleaned := filepath.Clean(path)
	suffix := make([]string, 0, 4)
	for {
		_, err := os.Lstat(cleaned)
		if err == nil {
			return cleaned, suffix
		}
		parent := filepath.Dir(cleaned)
		if parent == cleaned {
			return cleaned, suffix
		}
		suffix = append([]string{filepath.Base(cleaned)}, suffix...)
		cleaned = parent
	}
}

func isOutside(rel string) bool {
	return rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) || filepath.IsAbs(rel)
}
