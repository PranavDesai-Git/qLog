package chat

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)


var fileRegex = regexp.MustCompile(`\[\[(.*?)\]\]`)

func fileInject(input *string) {
	matches := fileRegex.FindAllStringSubmatch(*input, -1)
	if len(matches) == 0 {
		return
	}
	for _, match := range matches {
		tag := match[0]
		path := strings.TrimSpace(match[1])
		content, err := os.ReadFile(path)
		var replacement string
		if err != nil {
			replacement = fmt.Sprintf("\n FILE PATH: AT %s WAS UNABLE TO BE READ\nERR:\n%v\n", tag, err)
		} else {
			replacement = fmt.Sprintf("\n---- START OF FILE: %s ----\n%s\n---- END OF FILE ----", path, content)
		}
		*input = strings.Replace(*input, tag, replacement, 1)
	}
}

func BuildPrompt(input string) string {
	modInput := input
	fileInject(&modInput)
	return modInput
}
