package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type scenarioEvidence struct {
	SourceRoot        string         `json:"source_root"`
	SourceHead        string         `json:"source_head"`
	CandidateRevision string         `json:"candidate_revision"`
	Runner            string         `json:"runner"`
	Executor          string         `json:"executor"`
	BuildFlags        string         `json:"build_go_flags"`
	RunFlags          string         `json:"run_go_flags"`
	BuildDirectories  []string       `json:"build_directories"`
	Binaries          []candidate    `json:"binaries"`
	Runs              []*scenarioRun `json:"runs"`
}

type scenarioRun struct {
	Name        string          `json:"name"`
	Directory   string          `json:"directory"`
	Inputs      json.RawMessage `json:"inputs,omitempty"`
	Events      string          `json:"events_jsonl"` // 原始 JSONL 完整保留，故障留下的部分行也可读取。
	Log         string          `json:"go_test_log"`
	PassedTests []string        `json:"passed_tests"`
}

func (r *report) collectScenario(run *scenarioRun, tests []string) error {
	return r.step("scenario_"+run.Name+"_evidence", func(step *stepResult) error {
		evidence := filepath.Join(run.Directory, "scenario-evidence")
		inputs, inputErr := os.ReadFile(filepath.Join(evidence, "inputs.json"))
		step.Output = string(inputs)
		if inputErr == nil {
			if json.Valid(inputs) {
				run.Inputs = inputs
			} else {
				inputErr = errors.New("invalid scenario inputs JSON; original bytes are in step output")
			}
		}
		events, eventErr := os.ReadFile(filepath.Join(evidence, "events.jsonl"))
		run.Events = string(events)
		run.Log = filepath.Join(evidence, "go-test.log")
		log, logErr := os.ReadFile(run.Log)
		var assertions []error
		for _, test := range tests {
			if strings.Contains(string(log), "--- PASS: "+test+" (") {
				run.PassedTests = append(run.PassedTests, test)
			} else {
				assertions = append(assertions, fmt.Errorf("missing successful scenario test %s; log: %s", test, run.Log))
			}
		}
		if !strings.Contains(string(log), "Platform test cwd: "+r.Directory+"\n") {
			assertions = append(assertions, fmt.Errorf("scenario did not report the independent working directory %s", r.Directory))
		}
		return errors.Join(inputErr, eventErr, logErr, errors.Join(assertions...))
	})
}
