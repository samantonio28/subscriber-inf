// Package main — конвертер Cucumber JSON (godog) в Allure-results.
//
// BDD-тесты на godog пишут отчёт в формате Cucumber JSON. Allure-CLI строит
// отчёты только из allure-results, поэтому в CI между прогоном BDD-тестов и
// генерацией Allure-отчёта запускается этот конвертер: каждый сценарий из
// cucumber.json превращается в allure-result (со всеми шагами Gherkin).
//
// Использование:
//
//	go run ./tools/cucumber2allure -in test/e2e/bdd/cucumber.json -out allure-results
package main

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// cucumberFeature — элемент Cucumber JSON (уровень фичи).
type cucumberFeature struct {
	Name        string           `json:"name"`
	ID          string           `json:"id"`
	Keyword     string           `json:"keyword"`
	URI         string           `json:"uri"`
	Description string           `json:"description"`
	Tags        []cucumberTag    `json:"tags"`
	Elements    []cucumberDetail `json:"elements"`
}

// cucumberDetail — сценарий внутри фичи.
type cucumberDetail struct {
	Name        string         `json:"name"`
	ID          string         `json:"id"`
	Keyword     string         `json:"keyword"`
	Type        string         `json:"type"`
	Description string         `json:"description"`
	Steps       []cucumberStep `json:"steps"`
	Tags        []cucumberTag  `json:"tags"`
}

type cucumberStep struct {
	Name     string         `json:"name"`
	Keyword  string         `json:"keyword"`
	Line     int            `json:"line"`
	Duration int64          `json:"-"`
	Result   cucumberResult `json:"result"`
}

type cucumberResult struct {
	Status       string `json:"status"`
	Duration     int64  `json:"duration"` // наносекунды
	ErrorMessage string `json:"error_message,omitempty"`
}

type cucumberTag struct {
	Name string `json:"name"`
	Line int    `json:"line"`
}

// allureLabel — метка allure-результата.
type allureLabel struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// allureStep — шаг внутри allure-результата.
type allureStep struct {
	Name          string              `json:"name"`
	Status        string              `json:"status"`
	Stage         string              `json:"stage"`
	Start         int64               `json:"start"`
	Stop          int64               `json:"stop"`
	StatusDetails allureStatusDetails `json:"statusDetails"`
}

type allureStatusDetails struct {
	Message string `json:"message,omitempty"`
	Trace   string `json:"trace,omitempty"`
}

// allureResult — файл allure-results/<uuid>-result.json.
type allureResult struct {
	UUID          string              `json:"uuid"`
	HistoryID     string              `json:"historyId"`
	TestCaseID    string              `json:"testCaseId"`
	Name          string              `json:"name"`
	FullName      string              `json:"fullName"`
	Description   string              `json:"description"`
	Status        string              `json:"status"`
	Stage         string              `json:"stage"`
	StatusDetails allureStatusDetails `json:"statusDetails"`
	Start         int64               `json:"start"`
	Stop          int64               `json:"stop"`
	Labels        []allureLabel       `json:"labels"`
	Steps         []allureStep        `json:"steps"`
}

// mapStatus переводит статус Cucumber в статус Allure.
func mapStatus(status string) string {
	switch status {
	case "passed":
		return "passed"
	case "failed":
		return "failed"
	case "undefined", "ambiguous", "pending":
		return "broken"
	default:
		return "broken"
	}
}

func hashHex(parts ...string) string {
	h := md5.Sum([]byte(fmt.Sprint(parts)))
	return hex.EncodeToString(h[:])
}

func newUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func convert(features []cucumberFeature) []allureResult {
	var results []allureResult
	now := time.Now().UnixMilli()

	for _, f := range features {
		for _, sc := range f.Elements {
			if sc.Type != "scenario" && sc.Keyword != "Scenario" && sc.Keyword != "Scenario Outline" {
				continue
			}

			res := allureResult{
				UUID:        newUUID(),
				Name:        sc.Name,
				FullName:    f.Name + ": " + sc.Name,
				Status:      "passed",
				Stage:       "finished",
				TestCaseID:  hashHex("cucumber", f.URI, sc.Name),
				HistoryID:   hashHex("cucumber", f.Name, sc.Name),
				Description: sc.Description,
				Start:       now,
				Labels: []allureLabel{
					{Name: "feature", Value: f.Name},
					{Name: "story", Value: sc.Name},
					{Name: "suite", Value: f.Name},
					{Name: "package", Value: f.URI},
					{Name: "testClass", Value: filepath.Base(f.URI)},
					{Name: "framework", Value: "godog (Cucumber BDD)"},
					{Name: "language", Value: "go"},
					{Name: "parentSuite", Value: "BDD E2E (ЛР №5)"},
				},
			}
			for _, t := range f.Tags {
				res.Labels = append(res.Labels, allureLabel{Name: "tag", Value: t.Name})
			}

			var totalNs int64
			for _, st := range sc.Steps {
				step := allureStep{
					Name:   strings2Trim(st.Keyword, st.Name),
					Status: mapStatus(st.Result.Status),
					Stage:  "finished",
				}
				step.Start = now
				step.Stop = now + st.Result.Duration/1_000_000
				if st.Result.ErrorMessage != "" {
					step.StatusDetails = allureStatusDetails{Message: st.Result.ErrorMessage}
				}
				res.Steps = append(res.Steps, step)
				totalNs += st.Result.Duration
				if st.Result.Status == "failed" && res.Status != "failed" {
					res.Status = "failed"
					res.StatusDetails = allureStatusDetails{Message: st.Result.ErrorMessage}
				}
			}
			res.Stop = res.Start + totalNs/1_000_000
			if len(sc.Steps) == 0 {
				res.Status = "broken"
				res.StatusDetails = allureStatusDetails{Message: "no steps reported by godog"}
			}
			results = append(results, res)
		}
	}
	return results
}

func strings2Trim(keyword, name string) string {
	return keyword + name
}

func main() {
	in := flag.String("in", "test/e2e/bdd/cucumber.json", "путь к cucumber.json (godog)")
	out := flag.String("out", "allure-results", "каталог для allure-results")
	flag.Parse()

	data, err := os.ReadFile(*in)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cucumber2allure: read %s: %v\n", *in, err)
		os.Exit(1)
	}
	var features []cucumberFeature
	if err := json.Unmarshal(data, &features); err != nil {
		fmt.Fprintf(os.Stderr, "cucumber2allure: parse %s: %v\n", *in, err)
		os.Exit(1)
	}

	if err := os.MkdirAll(*out, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "cucumber2allure: mkdir %s: %v\n", *out, err)
		os.Exit(1)
	}

	results := convert(features)
	for _, r := range results {
		path := filepath.Join(*out, r.UUID+"-result.json")
		data, err := json.MarshalIndent(r, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "cucumber2allure: marshal result: %v\n", err)
			os.Exit(1)
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "cucumber2allure: write %s: %v\n", path, err)
			os.Exit(1)
		}
		fmt.Printf("cucumber2allure: %s [%s] -> %s\n", r.FullName, r.Status, filepath.Base(path))
	}
	fmt.Printf("cucumber2allure: %d сценариев конвертировано в %s\n", len(results), *out)
}
