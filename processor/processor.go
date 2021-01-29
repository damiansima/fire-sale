package processor

import (
	"errors"
	"fmt"
	"github.com/damiansima/fire-sale/util"
	log "github.com/sirupsen/logrus"
	"github.com/storozhukBM/verifier"
	"strconv"
	"strings"
)

const startMark = "{{"
const endMark = "}}"

func Process(original string) (string, error) {
	log.Tracef("Original string to parse: %s", original)

	parsedCopy := original
	placeholders := getPlaceholders(parsedCopy)
	placeholderMap, err := buildPlaceholderMap(placeholders)
	if err != nil {
		log.Warnf("Fail to process: %s", original)
		return "", err
	}

	processed := replace(original, placeholders, placeholderMap)
	return processed, nil
}

func replace(template string, placeholders []string, placeholderMap map[string]func() string) string {
	processed := template
	for _, placeholder := range placeholders {
		key := strings.Trim(placeholder, startMark)
		key = strings.Trim(key, endMark)
		processed = strings.Replace(processed, placeholder, placeholderMap[key](), 1)
	}
	log.Tracef("Replaced  template: %v", processed)
	return processed
}

func getPlaceholders(template string) []string {
	var placeholders []string

	keepParsing := true
	for keepParsing {
		startIdx := strings.Index(template, startMark)
		endIdx := strings.Index(template, endMark) + 2
		if startIdx < 0 {
			log.Tracef("Nothing found to replace")
			return []string{}
		}

		placeholder := template[startIdx:endIdx]

		log.Tracef("Found placeholder: %s", placeholder)
		placeholders = append(placeholders, placeholder)

		if len(template) > endIdx {
			template = template[endIdx:len(template)]
		} else {
			keepParsing = false
		}
	}
	log.Tracef("Placeholders found: %v", placeholders)
	return placeholders
}

func buildPlaceholderMap(placeholders []string) (map[string]func() string, error) {
	placeholderMap := make(map[string]func() string)
	for _, placeholder := range placeholders {
		placeholder = strings.Trim(placeholder, startMark)
		placeholder = strings.Trim(placeholder, endMark)

		name := placeholder[0:strings.Index(placeholder, "(")]
		strParams := placeholder[strings.Index(placeholder, "(")+1 : strings.Index(placeholder, ")")]
		params := strings.Split(strParams, ",")

		var err error
		placeholderMap[placeholder], err = funcBuilder(name, params)
		if err != nil {
			return placeholderMap, err
		}
	}
	return placeholderMap, nil
}

func funcBuilder(name string, params []string) (func() string, error) {
	if util.RandInRangeName == name {
		verify := verifier.New()
		verify.That(len(params) == 2, "The function %s takes at least %d parameters", name, 2)
		if verify.GetError() != nil {
			log.Warnf("Fail to parse function %s, error: %v", name, verify.GetError())
			return nil, verify.GetError()
		}
		param1, err := strconv.Atoi(strings.TrimSpace(params[0]))
		verify.That(err == nil, "The parameter  %s must be and int", param1)
		param2, err := strconv.Atoi(strings.TrimSpace(params[1]))
		verify.That(err == nil, "The parameter  %s must be and int", param2)
		if verify.GetError() != nil {
			log.Warnf("Fail to parse function %s, error: %v", name, verify.GetError())
			return nil, verify.GetError()
		}

		return func() string {
			return strconv.Itoa(util.RandInRange(param1, param2))
		}, nil
	}

	if util.RandInListName == name {
		verify := verifier.New()
		verify.That(len(params) >= 1, "The function %s takes at least %d parameters", name, 1)
		if verify.GetError() != nil {
			log.Warnf("Fail to parse function %s, error: %v", name, verify.GetError())
			return nil, verify.GetError()
		}
		return func() string {
			return util.RandInList(params)
		}, nil
	}

	if util.RandInFileName == name {
		verify := verifier.New()
		verify.That(len(params) == 1, "The function %s takes at least %d parameters", name, 1)
		if verify.GetError() != nil {
			log.Warnf("Fail to parse function %s, error: %v", name, verify.GetError())
			return nil, verify.GetError()
		}

		param1 := params[0]
		return func() string {
			return util.RandInFile(param1)
		}, nil
	}
	return nil, errors.New(fmt.Sprintf("Fail to parse function %s is not a recognized function name", name))
}
