package tag

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type WalkTagManager interface {
	ScanForTags(path string, file *os.File, info os.FileInfo) ([]Tag, error)
}

type WalkFileOperator interface {
	Open(fileName string) (*os.File, error)
	WalkDir(root string, fn fs.WalkDirFunc) error
}

type WalkParams struct {
	Root           string
	TagManager     WalkTagManager
	FileOperator   WalkFileOperator
	IgnorePatterns []GitIgnorePattern
}

func Walk(arg WalkParams) ([]Tag, error) {
	tags := make([]Tag, 0)

	err := arg.FileOperator.WalkDir(arg.Root, func(path string, d fs.DirEntry, wErr error) error {
		isValidPath := validatePath(path, arg.IgnorePatterns)

		if d.IsDir() {
			isGitDir := strings.Contains(d.Name(), ".git")

			if isGitDir || !isValidPath {
				return filepath.SkipDir
			}
			return nil
		}

		if !isValidPath {
			return nil
		}

		file, err := arg.FileOperator.Open(path)
		if err != nil {
			return err
		}

		fileInfo, err := d.Info()
		if err != nil {
			return err
		}

		foundTags, err := arg.TagManager.ScanForTags(path, file, fileInfo)
		if err != nil {
			return err
		}

		tags = append(tags, foundTags...)

		if closeErr := file.Close(); closeErr != nil {
			return closeErr
		}

		return wErr
	})

	return tags, err
}

func validatePath(path string, ignorePatterns []GitIgnorePattern) bool {
	for _, v := range ignorePatterns {
		matched := v.Match([]byte(path))
		if matched {
			return false
		}
	}
	return true
}
