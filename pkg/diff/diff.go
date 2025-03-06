package diff

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GetGitDiff retrieves the git diff against the specified branch.
func GetGitDiff(branch string) (string, error) {
	cmd := exec.Command("git", "diff", branch)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error executing git diff: %v, %s", err, stderr.String())
	}
	return out.String(), nil
}

// GetChangedFiles returns a slice of filenames that have been changed or added compared to the specified branch.
// If includeUntracked is true, it also includes untracked files.
func GetChangedFiles(branch string, includeUntracked bool) ([]string, error) {
	// Get tracked files with changes
	cmd := exec.Command("git", "diff", "--name-only", branch)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("error getting changed files: %v, %s", err, stderr.String())
	}
	files := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(files) == 1 && files[0] == "" {
		files = []string{}
	}

	// If includeUntracked is true, also add untracked files
	if includeUntracked {
		untrackedFiles, err := GetUntrackedFiles()
		if err != nil {
			return nil, fmt.Errorf("error getting untracked files: %v", err)
		}
		files = append(files, untrackedFiles...)
	}

	return files, nil
}

// GetUntrackedFiles returns a slice of all untracked files, including files in untracked directories.
func GetUntrackedFiles() ([]string, error) {
	// Get all untracked files, including those in untracked directories
	cmd := exec.Command("git", "ls-files", "--others", "--exclude-standard")
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("error getting untracked files: %v, %s", err, stderr.String())
	}

	files := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(files) == 1 && files[0] == "" {
		return []string{}, nil
	}

	return files, nil
}

// GetFileContent retrieves the content of a file from the specified branch using git show.
func GetFileContent(branch, filename string) (string, error) {
	cmd := exec.Command("git", "show", fmt.Sprintf("%s:%s", branch, filename))
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error getting file content for %s from branch %s: %v, %s", filename, branch, err, stderr.String())
	}
	return out.String(), nil
}

// GetLocalFileContent retrieves the content of a file from the local filesystem.
func GetLocalFileContent(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("error reading local file %s: %v", filename, err)
	}
	return string(content), nil
}

// IsDirectory checks if a given path is a directory
func IsDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.IsDir(), nil
}

// GetAllFilesInDirectory recursively gets all files in a directory
func GetAllFilesInDirectory(dirPath string) ([]string, error) {
	var files []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip .git directory
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		// Only add files, not directories
		if !info.IsDir() {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory %s: %v", dirPath, err)
	}

	return files, nil
}
