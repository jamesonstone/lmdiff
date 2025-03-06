package prompt

import (
	"fmt"
	"strings"
)

// ConstructLLMPrompt creates an XML formatted LLM prompt with the git diff and original file contents.
// It now includes all changed and added files and instructs to review best practices for all files.
func ConstructLLMPrompt(diff string, changedFiles []string, files map[string]string) string {
	var promptBuilder strings.Builder

	promptBuilder.WriteString("<prompt>\n")
	promptBuilder.WriteString(". <description>")
	promptBuilder.WriteString("Please analyze the git diff changes. Review the best practices of all files, including new files. Please use KISS+YAGNI+DRY+SOLID principles.")
	promptBuilder.WriteString("Assess the new changes against existing files, suggest improvements, and ask clarifying questions if needed. Complete your review by providing a summary of the changes in paragraph form followed by a bulleted list of suggested changes.")
	promptBuilder.WriteString("</description>\n")
	promptBuilder.WriteString("  <changedFiles>\n")
	for _, filename := range changedFiles {
		if strings.TrimSpace(filename) == "" {
			continue
		}
		promptBuilder.WriteString(fmt.Sprintf("    <file name=\"%s\"/>\n", filename))
	}
	promptBuilder.WriteString("  </changedFiles>\n")
	promptBuilder.WriteString("  <files>\n")
	for filename, content := range files {
		promptBuilder.WriteString(fmt.Sprintf("    <file name=\"%s\">\n", filename))
		promptBuilder.WriteString("      <![CDATA[\n")
		promptBuilder.WriteString(content)
		promptBuilder.WriteString("      ]]>\n")
		promptBuilder.WriteString("    </file>\n")
	}
	promptBuilder.WriteString("  </files>\n")
	promptBuilder.WriteString("  <gitDiff>\n")
	promptBuilder.WriteString("    <![CDATA[\n")
	promptBuilder.WriteString(diff)
	promptBuilder.WriteString("    ]]>\n")
	promptBuilder.WriteString("  </gitDiff>\n")
	promptBuilder.WriteString("</prompt>\n")

	return promptBuilder.String()
}
