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
	ID                string   `yaml:"id" json:"test_number"`
	Text              string   `json:"test_desc"`
	Audit             string   `json:"audit"`
	AuditEnv          string   `yaml:"audit_env"`
	AuditConfig       string   `yaml:"audit_config"`
	Type              string   `json:"type"`
	Tests             *tests   `json:"-"`
	Set               bool     `json:"-"`
	Remediation       string   `json:"remediation"`
	TestInfo          []string `json:"test_info"`
	State             `json:"status"`
	ActualValue       string `json:"actual_value"`
	Scored            bool   `json:"scored"`
	IsMultiple        bool   `yaml:"use_multiple_values"`
	ExpectedResult    string `json:"expected_result"`
	Reason            string `json:"reason,omitempty"`
	AuditOutput       string `json:"-"`
	AuditEnvOutput    string `json:"-"`
	AuditConfigOutput string `json:"-"`
	DisableEnvTesting bool   `json:"-"`
}

type binOp string

const (
	and                   binOp = "and"
	or                          = "or"
	defaultArraySeparator       = ","
)

type tests struct {
	TestItems []*testItem `yaml:"test_items"`
	BinOp     binOp       `yaml:"bin_op"`
}

type AuditUsed string

const (
	AuditCommand AuditUsed = "auditCommand"
	AuditConfig  AuditUsed = "auditConfig"
	AuditEnv     AuditUsed = "auditEnv"
)

type testItem struct {
	Flag             string
	Env              string
	Path             string
	Output           string
	Value            string
	Set              bool
	Compare          compare
	isMultipleOutput bool
	auditUsed        AuditUsed
}

type compare struct {
	Op    string
	Value string
}

type testOutput struct {
	testResult     bool
	flagFound      bool
	actualResult   string
	ExpectedResult string
}

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
	// for {
	// 	err := decoder.Decode(&body)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	if err == io.EOF {
	// 		break
	// 	}
	// }

	//fmt.Println(body.Controls[0].ID)
	//fmt.Println(body.Controls[1].Summary) //not showing correct output
	// fmt.Println(body.Controls[0].Tests[0].Fail)
	// fmt.Println(body.Controls[0].Tests[0].Results[0].Status)
	// fmt.Println(body.Totals.TotalPass) // not showing correct output

	// Prints entire json as yaml (Caution: A few fields are buggy, to be fixed)
	// y, err := yaml.Marshal(body)
	// if err != nil {
	// 	fmt.Printf("err: %v\n", err)
	// 	return
	// }
	// fmt.Println(string(y))
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
			Name: "sample-policy-report",
		},
	}

	// Create Deployment
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
