package git

import (
	"fmt"
	"strings"

	"omar-kada/air-compose/models"

	gitObject "github.com/go-git/go-git/v6/plumbing/object"
)

// PatchParser is an interface for parsing git patches into structured data.
type PatchParser interface {
	Parse(patch string, commit *gitObject.Commit) (Patch, error)
}

type parser struct{}

// NewPatchParser creates a new instance of a PatchParser.
func NewPatchParser() PatchParser {
	return parser{}
}

// Parse converts a git patch and commit into a structured Patch object.
// It extracts file-level diffs from the patch, associates them with the commit metadata,
// and returns a Patch containing all this information.
func (parser) Parse(diff string, commit *gitObject.Commit) (Patch, error) {
	// Split the diff string into separate file diffs based on "diff --git" markers
	diffsByFile := strings.Split(diff, "diff --git")

	// Remove the first empty element if it exists
	if len(diffsByFile) > 0 && diffsByFile[0] == "" {
		diffsByFile = diffsByFile[1:]
	}
	var fileDiffs []models.FileDiff

	for _, diffStr := range diffsByFile {
		fileDiff, err := toFileDiff("diff --git" + diffStr)
		if err != nil {
			return Patch{}, err
		}
		fileDiffs = append(fileDiffs, fileDiff)
	}
	return Patch{
		Title:      commit.Message,
		Diff:       diff,
		Files:      fileDiffs,
		Author:     commit.Author.Name,
		CommitHash: commit.Hash.String(),
	}, nil
}

func toFileDiff(strDiff string) (models.FileDiff, error) {
	parts := strings.SplitN(strDiff, "\n", 2)
	if len(parts) < 2 {
		return models.FileDiff{}, fmt.Errorf("diff contains less than 2 lines")
	}
	header := parts[0]
	fileNames := strings.Fields(header)
	if len(fileNames) <= 3 {
		return models.FileDiff{}, fmt.Errorf("can't find file names")
	}
	// extract file names while removing a/... and b/...
	oldFile := fileNames[2][2:]
	newFile := fileNames[3][2:]
	return models.FileDiff{
		OldFile: oldFile,
		NewFile: newFile,
		Diff:    strings.TrimSuffix(strDiff, "\n"),
	}, nil
}
