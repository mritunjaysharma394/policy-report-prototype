// package
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	appsv1aplha1 "github.com/mritunjaysharma394/policy-report-prototype/pkg/apis/wgpolicyk8s.io/v1alpha1"
	"k8s.io/client-go/tools/clientcmd"

	client "github.com/mritunjaysharma394/policy-report-prototype/pkg/generated/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"strconv"

	"k8s.io/client-go/util/homedir"
)

type OverallControls struct {
	Controls []*Controls
	Totals   Summary
}

var body OverallControls

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
	ID                string   `yaml:"id" json:"test_number"`
	Text              string   `json:"test_desc"`
	Audit             string   `json:"audit"`
	AuditEnv          string   `yaml:"audit_env"`
	AuditConfig       string   `yaml:"audit_config"`
	Type              string   `json:"type"`
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

// func runDockerImage() error {
// 	if err := exec.Command("docker", "run", "-rm", ".v", "`pwd`", "host", "aquasec/kube-bench:latest install").Run(); err != nil {
// 		log.Fatal(err)
// 		return err
// 	}
// 	return nil
// }

func runKubeBench(jsonPath string) (string, error) {

	// err := runDockerImage()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//executes kube-bench
	// cmd := exec.Command("./kube-bench", "--benchmark", "gke-1.0", "--json")
	// out, err := cmd.CombinedOutput()
	out, err := ioutil.ReadFile(jsonPath)
	return string(out), err
}

func getBody(jsonPath string) (*OverallControls, error) {
	//calls function that runs KubeBench
	out, err := runKubeBench(jsonPath)

	if err != nil {
		log.Fatal(err)
	}

	jsonDataReader := strings.NewReader(out)
	decoder := json.NewDecoder(jsonDataReader)

	err = decoder.Decode(&body)
	return &body, err
}

func getArguments() (string, string, string, string, *string) {
	var jsonPath, policyName, namespace, category string
	flag.StringVar(&jsonPath, "jsonPath", "check.json", "Path to the JSON file")
	flag.StringVar(&policyName, "policyName", "", "name of policy report")
	flag.StringVar(&namespace, "namespace", "default", "namespace of the cluster")
	flag.StringVar(&category, "category", "CIS Benchmarks for Kubernetes", "category of the policy report")

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.Parse()
	return jsonPath, policyName, namespace, category, kubeconfig
}

func getPolicyReportsResult(category string, control *Controls, group *Group, check *Check) *appsv1aplha1.PolicyReportResult {
	Result := appsv1aplha1.PolicyReportResult{
		Policy:      control.Text,
		Rule:        group.Text,
		Category:    category,
		Result:      strings.ToLower(string(check.State)),
		Scored:      check.Scored,
		Description: check.Text,
		Properties: map[string]string{
			"index":           check.ID,
			"audit":           check.Audit,
			"AuditEnv":        check.AuditEnv,
			"AuditConfig":     check.AuditConfig,
			"type":            check.Type,
			"remediation":     check.Remediation,
			"test_info":       check.TestInfo[0],
			"actual_value":    check.ActualValue,
			"IsMultiple":      strconv.FormatBool(check.IsMultiple),
			"expected_result": check.ExpectedResult,
			"reason":          check.Reason,
		},
	}
	return &Result
}

func createPolicyReport(jsonPath string, policyName string, namespace string, category string, kubeconfig *string, policy *appsv1aplha1.PolicyReport) error {

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := client.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	body, err := getBody(jsonPath)
	if err != nil {
		panic(err)
	}

	policy = &appsv1aplha1.PolicyReport{
		ObjectMeta: metav1.ObjectMeta{
			Name: policyName,
		},
		Summary: appsv1aplha1.PolicyReportSummary{
			Pass: body.Totals.Pass,
			Fail: body.Totals.Fail,
			Warn: body.Totals.Warn,
		},
	}

	for _, control := range body.Controls {
		for _, group := range control.Groups {
			for _, check := range group.Checks {
				_ = check
				policy.Results = append(policy.Results, getPolicyReportsResult(category, control, group, check))
			}
		}
	}

	policyReports := clientset.Wgpolicyk8sV1alpha1().PolicyReports(namespace)
	// Create Policy-Report
	fmt.Println("Creating policy-report...")
	result, err := policyReports.Create(context.TODO(), policy, metav1.CreateOptions{})
	if err != nil {
		return (err)
	}
	fmt.Printf("Created policy-report %q.\n", result.GetObjectMeta().GetName())
	return nil
}

func main() {

	jsonPath, policyName, namespace, category, kubeconfig := getArguments()
	var policy *appsv1aplha1.PolicyReport
	cmdStr := "docker run --rm -v `pwd`:/host aquasec/kube-bench:latest install"
	exec.Command("/bin/sh", "-c", cmdStr).Run()
	err := createPolicyReport(jsonPath, policyName, namespace, category, kubeconfig, policy)
	if err != nil {
		log.Fatal(err)
	}
}
