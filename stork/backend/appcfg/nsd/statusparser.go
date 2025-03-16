package nsdconfig

import (
	"errors"
	"regexp"
)

func ParseNSDVersion(input string) (string, error) {
	version := regexp.MustCompile(`version:\s*(\d+\.\d+\.\d+)(?:-(?:[a-zA-Z0-9]+))?`)
	matches := version.FindStringSubmatch(input)
	if len(matches) == 0 {
		return "", errors.New("version not found")
	}
	return matches[1], nil
}
