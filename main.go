// package
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"k8s.io/client-go/tools/clientcmd"
	appsv1aplha1 "k8s.io/sample-controller/pkg/apis/wgpolicyk8s.io/v1alpha1"

	client "github.com/policy-report-prototype/pkg/generated/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"strconv"

	"k8s.io/client-go/util/homedir"
)

type OverallControls struct {
	Controls []*Controls
	Totals   Summary
}

// Controls holds all controls to check for master nodes.
type Controls struct {
	ID      string   `yaml:"id" json:"id"`
	Version string   `json:"version"`
	Text    string   `json:"text"`
	Type    NodeType `json:"node_type"`
	Groups  []*Group `json:"tests"`
	Summary
}

// Group is a collection of similar checks.
type Group struct {
	ID     string   `yaml:"id" json:"section"`
	Type   string   `yaml:"type" json:"type"`
	Pass   int      `json:"pass"`
	Fail   int      `json:"fail"`
	Warn   int      `json:"warn"`
	Info   int      `json:"info"`
	Text   string   `json:"desc"`
	Checks []*Check `json:"results"`
}

// Summary is a summary of the results of control checks run.
type Summary struct {
	Pass int `json:"total_pass"`
	Fail int `json:"total_fail"`
	Warn int `json:"total_warn"`
	Info int `json:"total_info"`
}

// Predicate a predicate on the given Group and Check arguments.
type Predicate func(group *Group, check *Check) bool

// NodeType indicates the type of node (master, node).
type NodeType string

// State is the state of a control check.
type State string

const (
	// PASS check passed.
	PASS State = "PASS"
	// FAIL check failed.
	FAIL State = "FAIL"
	// WARN could not carry out check.
	WARN State = "WARN"
	// INFO informational message
	INFO State = "INFO"

	// SKIP for when a check should be skipped.
	SKIP = "skip"

	// MASTER a master node
	MASTER NodeType = "master"
	// NODE a node
	NODE NodeType = "node"
	// FEDERATED a federated deployment.
	FEDERATED NodeType = "federated"

	// ETCD an etcd node
	ETCD NodeType = "etcd"
	// CONTROLPLANE a control plane node
	CONTROLPLANE NodeType = "controlplane"
	// POLICIES a node to run policies from
	POLICIES NodeType = "policies"
	// MANAGEDSERVICES a node to run managedservices from
	MANAGEDSERVICES = "managedservices"

	// MANUAL Check Type
	MANUAL string = "manual"
)

// Check contains information about a recommendation in the
// CIS Kubernetes document.
type Check struct {
	ID          string `yaml:"id" json:"test_number"`
	Text        string `json:"test_desc"`
	Audit       string `json:"audit"`
	AuditEnv    string `yaml:"audit_env"`
	AuditConfig string `yaml:"audit_config"`
	Type        string `json:"type"`
	// Tests             *tests   `json:"-"`
	Set               bool     `json:"-"`
	Remediation       string   `json:"remediation"`
	TestInfo          []string `json:"test_info"`
	State             string   `json:"status"`
	ActualValue       string   `json:"actual_value"`
	Scored            bool     `json:"scored"`
	IsMultiple        bool     `yaml:"use_multiple_values"`
	ExpectedResult    string   `json:"expected_result"`
	Reason            string   `json:"reason,omitempty"`
	AuditOutput       string   `json:"-"`
	AuditEnvOutput    string   `json:"-"`
	AuditConfigOutput string   `json:"-"`
	DisableEnvTesting bool     `json:"-"`
}

type AuditUsed string

const (
	AuditCommand AuditUsed = "auditCommand"
	AuditConfig  AuditUsed = "auditConfig"
	AuditEnv     AuditUsed = "auditEnv"
)

func main() {

	//calls function that runs KubeBench
	out := runKubeBench()
	jsonDataReader := strings.NewReader(out)
	decoder := json.NewDecoder(jsonDataReader)

	var body OverallControls
	err := decoder.Decode(&body)
	if err != nil {
		panic(err)
	}
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := client.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	ats := clientset.Wgpolicyk8sV1alpha1().PolicyReports("default")
	deployment := &appsv1aplha1.PolicyReport{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cis-dummy-policy",
		},
		Summary: appsv1aplha1.PolicyReportSummary{
			Pass: body.Totals.Pass,
			Fail: body.Totals.Fail,
			Warn: body.Totals.Warn,
		},
	}

	for i := 0; i < 5; i++ {
		for j := 0; j < len(body.Controls[i].Groups); j++ {
			for k := 0; k < len(body.Controls[i].Groups[j].Checks); k++ {
				deployment.Results = append(deployment.Results, test_out(body, i, j, k))
			}
		}
	}
	// Create Policy-Report
	fmt.Println("Creating policy-report...")
	result, err := ats.Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created policy-report %q.\n", result.GetObjectMeta().GetName())

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

func test_out(body OverallControls, i int, j int, k int) *appsv1aplha1.PolicyReportResult {
	Result := appsv1aplha1.PolicyReportResult{
		Policy:      body.Controls[i].Text,
		Rule:        body.Controls[i].Groups[j].Text,
		Category:    body.Controls[i].Groups[j].Text,
		Result:      strings.ToLower(string(body.Controls[i].Groups[j].Checks[k].State)),
		Scored:      body.Controls[i].Groups[j].Checks[k].Scored,
		Description: body.Controls[i].Groups[j].Checks[k].Text,
		Properties: map[string]string{
			"index":           body.Controls[i].Groups[j].Checks[k].ID,
			"audit":           body.Controls[i].Groups[j].Checks[k].Audit,
			"AuditEnv":        body.Controls[i].Groups[j].Checks[k].AuditEnv,
			"AuditConfig":     body.Controls[i].Groups[j].Checks[k].AuditConfig,
			"type":            body.Controls[i].Groups[j].Checks[k].Type,
			"remediation":     body.Controls[i].Groups[j].Checks[k].Remediation,
			"test_info":       body.Controls[i].Groups[j].Checks[k].TestInfo[0],
			"actual_value":    body.Controls[i].Groups[j].Checks[k].ActualValue,
			"IsMultiple":      strconv.FormatBool(body.Controls[i].Groups[j].Checks[k].IsMultiple),
			"expected_result": body.Controls[i].Groups[j].Checks[k].ExpectedResult,
			"reason":          body.Controls[i].Groups[j].Checks[k].Reason,
		},
	}
	return &Result
}
