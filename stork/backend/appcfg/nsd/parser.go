package nsdconfig

import (
	"bufio"
	"os"
	"strings"
)

type Parser struct {
	filename string
}

type ParsedValue struct {
	StringValue *string
	BoolValue   *bool
}

func NewParser(filename string) *Parser {
	return &Parser{
		filename,
	}
}

func (p *Parser) Parse() (map[string]ParsedValue, error) {
	file, err := os.Open(p.filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	values := make(map[string]ParsedValue)
	for scanner.Scan() {
		text := strings.Trim(scanner.Text(), " ")
		if !strings.HasPrefix(text, "#") {
			split := strings.Split(text, ":")
			for i := range split {
				split[i] = strings.Trim(split[i], " ")
			}
			switch len(split) {
			case 0:
				continue
			case 1:
				var boolValue = true
				values[split[0]] = ParsedValue{
					BoolValue: &boolValue,
				}
			default:
				switch split[1] {
				case "yes":
					var boolValue = true
					values[split[0]] = ParsedValue{
						BoolValue: &boolValue,
					}
				case "no":
					var boolValue = false
					values[split[0]] = ParsedValue{
						BoolValue: &boolValue,
					}
				default:
					values[split[0]] = ParsedValue{
						StringValue: &split[1],
					}
				}
			}
		}
	}
	return values, nil
}
