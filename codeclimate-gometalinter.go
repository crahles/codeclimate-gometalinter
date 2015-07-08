package main

import (
	"regexp"

	"github.com/codeclimate/cc-engine-go/engine"
)
import "strings"
import "os"
import "os/exec"
import "strconv"
import "sort"

func main() {
	rootPath := "/code/"
	analysisFiles, err := engine.GoFileWalk(rootPath)
	if err != nil {
		os.Exit(1)
	}

	config, err := engine.LoadConfig()
	if err != nil {
		os.Exit(1)
	}

	excludedFiles := []string{}
	if config["exclude_paths"] != nil {
		for _, file := range config["exclude_paths"].([]interface{}) {
			excludedFiles = append(excludedFiles, file.(string))
		}
		sort.Strings(excludedFiles)
	}

	for _, path := range analysisFiles {
		relativePath := strings.SplitAfter(path, rootPath)[1]
		i := sort.SearchStrings(excludedFiles, relativePath)
		if i < len(excludedFiles) && excludedFiles[i] == relativePath {
			continue
		}

		cmd := exec.Command("gometalinter", path)
		out, err := cmd.CombinedOutput()

		if err != nil && err.Error() != "exit status 1" {
			// engine.PrintWarning()
			return
		}

		lines := strings.Split(string(out), "\n")
		// fmt.Println(lines)
		if len(lines) > 1 {
			for _, line := range lines[:len(lines)-1] {

				pieces := strings.SplitN(line, ":", 5)
				if len(pieces) < 3 {
					// engine.PrintWarning()
					return
				}
				lineNo, err := strconv.Atoi(pieces[1])
				if err != nil {
					// engine.PrintWarning()
					return
				}
				colNo, err := strconv.Atoi(pieces[2])
				if err != nil {
					// engine.PrintWarning()
					return
				}

				re := regexp.MustCompile(`\(([^\)]*)\)$`)

				m := re.FindAllString(pieces[4], 1)
				check := m[0]
				check = "gometalinter/" + check[1:len(check)-1]

				desc := re.ReplaceAllString(pieces[4], "")
				desc = strings.TrimSpace(desc)

				var cat []string
				switch check {
				case "deadcode", "ineffassign":
					cat = []string{"Clarity"}
				case "golint", "gotype":
					cat = []string{"Style"}
				case "defercheck", "dupl":
					cat = []string{"Duplication"}
				case "gocyclo":
					cat = []string{"Complexity"}
				default:
					cat = []string{"Bug Risk"}
				}

				issue := &engine.Issue{
					Type:        "issue",
					Check:       check,
					Description: desc,
					// RemediationPoints: 500,
					Categories: cat,
					Location: &engine.Location{
						Path: strings.SplitAfter(path, rootPath)[1],
						Positions: &engine.LineColumnPosition{
							Begin: &engine.LineColumn{
								Line:   lineNo,
								Column: colNo,
							},
						},
					},
				}
				engine.PrintIssue(issue)
			}
		}
	}

}
