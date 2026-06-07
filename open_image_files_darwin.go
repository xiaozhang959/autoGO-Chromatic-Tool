//go:build darwin
// +build darwin

package main

import (
	"fmt"
	"os/exec"
	"strings"

	nativedialog "github.com/sqweek/dialog"
)

func openImageFiles() ([]string, error) {
	script := `
set selectedFiles to choose file with prompt "选择图片文件" of type {"png", "jpg", "jpeg", "bmp"} with multiple selections allowed
set outputPaths to {}
repeat with selectedFile in selectedFiles
	set end of outputPaths to POSIX path of selectedFile
end repeat
set oldDelimiters to AppleScript's text item delimiters
set AppleScript's text item delimiters to linefeed
set outputText to outputPaths as text
set AppleScript's text item delimiters to oldDelimiters
return outputText
`

	output, err := exec.Command("osascript", "-e", script).CombinedOutput()
	if err != nil {
		text := strings.TrimSpace(string(output))
		if strings.Contains(text, "(-128)") {
			return nil, nativedialog.ErrCancelled
		}
		if text == "" {
			return nil, err
		}
		return nil, fmt.Errorf("%s", text)
	}

	filePaths := splitAppleScriptLines(string(output))
	if len(filePaths) == 0 {
		return nil, nativedialog.ErrCancelled
	}
	return filePaths, nil
}

func splitAppleScriptLines(output string) []string {
	lines := strings.Split(strings.TrimRight(output, "\r\n"), "\n")
	filePaths := make([]string, 0, len(lines))
	for _, line := range lines {
		filePath := strings.TrimRight(line, "\r")
		if filePath == "" {
			continue
		}
		filePaths = append(filePaths, filePath)
	}
	return filePaths
}
