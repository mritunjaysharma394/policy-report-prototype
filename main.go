// package
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"sigs.k8s.io/yaml"
)

type Body struct {
	Controls []Control `json:"Controls,omitempty"`
	Totals   Total     `json:"Totals,omitempty"`
}

type Control struct {
	ID        string `json:"id,omitempty"`
	Version   string `json:"version,omitempty"`
	Text      string `json:"text,omitempty"`
	NodeType  string `json:"node_type,omitempty"`
	Tests     []Test `json:"tests,omitempty"`
	TotalPass int    `json: "total_pass,omitempty"`
	TotalFail int    `json: "total_fail,omitempty"`
	TotalWarn int    `json: "total_warn,omitempty"`
	TotalInfo int    `json: "total_info,omitempty"`
}

type Test struct {
	Section string   `json: "section,omitempty"`
	Type    string   `json: "type,omitempty"`
	Pass    int      `json: "pass,omitempty"`
	Fail    int      `json: "fail,omitempty"`
	Warn    int      `json: "warn,omitempty"`
	Info    int      `json: "info,omitempty"`
	Desc    string   `json: "desc,omitempty"`
	Results []Result `json: "results,omitempty"`
}

type Result struct {
	TestNumber  string `json: "test_number"`
	TestDesc    string `json: "test_desc,omitempty"`
	Audit       string `json: "audit,omitempty"`
	AuditEnv    string `json: "AuditEnv,omitempty"`
	AuditConfig string `json: "AuditConfig,omitempty"`
	Type        string `json: "type,omitempty"`
	Remediation string `json: "remediation, omitempty"`
	// test_info has to be fixed
	//TestInfo []TestInformation `json: "test_info,omitempty"`
	Status         string `json: "status,omitempty"`
	ActualValue    string `json: "actual_value,omitempty"`
	Scored         bool   `json: "scored,omitempty"`
	IsMultiple     bool   `json: "IsMultiple,omitempty"`
	ExpectedResult string `json: "expected_result,omitempty"`
}

type Total struct {
	TotalPass int `json: "total_pass,omitempty"`
	TotalFail int `json: "total_fail,omitempty"`
	TotalWarn int `json: "total_warn,omitempty"`
	TotalInfo int `json: "total_info,omitempty"`
}

func main() {

	//calls function that runs KubeBench
	out := runKubeBench()
	jsonDataReader := strings.NewReader(out)
	decoder := json.NewDecoder(jsonDataReader)

	var body Body
	err := decoder.Decode(&body)
	if err != nil {
		panic(err)
	}
	// for {
	// 	err := decoder.Decode(&body)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	if err == io.EOF {
	// 		break
	// 	}
	// }

	fmt.Println(body.Controls[0].ID)
	fmt.Println(body.Controls[1].TotalPass) //not showing correct output
	fmt.Println(body.Controls[0].Tests[0].Fail)
	fmt.Println(body.Controls[0].Tests[0].Results[0].Status)
	fmt.Println(body.Totals.TotalPass) // not showing correct output

	// Prints entire json as yaml (Caution: A few fields are buggy, to be fixed)
	y, err := yaml.Marshal(body)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	fmt.Println(string(y))
}

func runKubeBench() string {

	//executes kube-bench
	cmd := exec.Command("./kube-bench", "--json")
	out, err := cmd.CombinedOutput()

	if err != nil {
		log.Fatal(err)
	}

	return string(out)
}
