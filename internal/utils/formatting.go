package utils

import (
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// show a unified diff of the two strings
func ShowDiff(old, new string) string {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(old, new, false)

	var result strings.Builder
	for _, diff := range diffs {
		switch diff.Type {
		case diffmatchpatch.DiffEqual:
			result.WriteString(" " + diff.Text)
		case diffmatchpatch.DiffDelete:
			result.WriteString("-" + diff.Text)
		case diffmatchpatch.DiffInsert:
			result.WriteString("+" + diff.Text)
		}
	}

	return result.String()
}
