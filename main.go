/**
 * @file main.go
 * @description
 * This is the main entry point for the Go application that packages git diff changes and original file contents
 * from a specified remote branch into an LLM reasoning prompt. The application reads the git diff and the content of
 * the files as they exist in the remote branch, constructs a prompt, and outputs it to stdout.
 *
 * Key Features:
 * - Parses command line argument for branch name (default is "main")
 * - Retrieves git diff using 'git diff' command
 * - Retrieves list of changed files using 'git diff --name-only'
 * - For each changed file, retrieves the original file content from the specified branch using 'git show'
 * - Constructs an LLM reasoning prompt including overall diff and original file contents
 * - Outputs the prompt for further assessment by a large language model
 *
 * @dependencies
 * - os/exec: to execute git commands
 * - flag: to parse command line arguments
 * - fmt, bytes, strings, log: for various utilities and error handling
 *
 * @notes
 * - Assumes that git is installed and that the application is run within a valid git repository.
 * - If a file does not exist in the specified branch, a warning is logged and a placeholder message is used.
 * - Error handling is implemented to gracefully handle command execution failures.
 */

package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"

	"github.com/jamesonstone/lmdiff/pkg/diff"
	"github.com/jamesonstone/lmdiff/pkg/prompt"
)

func main() {
	branch := flag.String("branch", "main", "The branch name to compare changes with (default: main)")
	copyFlag := flag.Bool("copy", false, "Automatically copy the prompt output to the clipboard")
	shortCopyFlag := flag.Bool("c", false, "Automatically copy the prompt output to the clipboard (shorthand)")
	includeUntrackedFlag := flag.Bool("include-untracked", true, "Include untracked files in the analysis (default: true)")
	flag.Parse()

	// Get the git diff relative to the specified branch.
	gitDiff, err := diff.GetGitDiff(*branch)
	if err != nil {
		log.Fatalf("Failed to get git diff: %v", err)
	}

	// Get the list of changed (including added) files compared to the specified branch.
	changedFiles, err := diff.GetChangedFiles(*branch, *includeUntrackedFlag)
	if err != nil {
		log.Fatalf("Failed to get list of changed files: %v", err)
	}

	// Map to hold original file contents for each changed file.
	originalFiles := make(map[string]string)

	// Process each file in the changedFiles list
	for _, file := range changedFiles {
		if file == "" {
			continue
		}

		// Check if this is a directory
		isDir, err := diff.IsDirectory(file)
		if err != nil {
			log.Printf("Warning: could not determine if %s is a directory: %v", file, err)
			continue
		}

		if isDir {
			// If it's a directory, get all files inside it
			dirFiles, err := diff.GetAllFilesInDirectory(file)
			if err != nil {
				log.Printf("Warning: could not read directory %s: %v", file, err)
				continue
			}

			// Process each file in the directory
			for _, dirFile := range dirFiles {
				content := processFile(dirFile, *branch)
				originalFiles[dirFile] = content
			}
		} else {
			// It's a regular file
			content := processFile(file, *branch)
			originalFiles[file] = content
		}
	}

	// Construct the final prompt.
	promptText := prompt.ConstructLLMPrompt(gitDiff, changedFiles, originalFiles)

	// Output the constructed prompt.
	fmt.Println(promptText)

	// If --copy or -c flag is set, copy the prompt output to the clipboard using pbcopy.
	shouldCopy := *copyFlag || *shortCopyFlag
	if shouldCopy {
		cmd := exec.Command("pbcopy")
		in, err := cmd.StdinPipe()
		if err != nil {
			log.Fatalf("Failed to get stdin pipe for pbcopy: %v", err)
		}
		if err := cmd.Start(); err != nil {
			log.Fatalf("Failed to start pbcopy: %v", err)
		}
		in.Write([]byte(promptText))
		in.Close()
		cmd.Wait()
		fmt.Println("Prompt copied to clipboard.")
	}
}

// processFile attempts to get the content of a file from either the branch or locally
func processFile(file string, branch string) string {
	content, err := diff.GetFileContent(branch, file)
	if err != nil {
		// For new files not present in the remote branch, read from local filesystem.
		localContent, err2 := diff.GetLocalFileContent(file)
		if err2 != nil {
			log.Printf("Warning: cannot retrieve content for %s: %v", file, err2)
			return "Error retrieving file content."
		} else {
			return localContent
		}
	}
	return content
}
