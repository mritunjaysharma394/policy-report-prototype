// package
package main

import (
	"path/filepath"
	"testing"

	appsv1aplha1 "github.com/mritunjaysharma394/policy-report-prototype/pkg/apis/wgpolicyk8s.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/homedir"
)

func TestCreatePolicyReport(t *testing.T) {
	policyTests := []struct {
		name         string
		policyreport *appsv1aplha1.PolicyReport
		ns           string
		category     string
	}{
		{"demo-test-1", &appsv1aplha1.PolicyReport{
			ObjectMeta: metav1.ObjectMeta{
				Name: "demo-1",
			},
			Summary: appsv1aplha1.PolicyReportSummary{
				Pass: 10,
				Fail: 4,
				Warn: 0,
			},
		}, "default", ""},
		{"demo-test-2", &appsv1aplha1.PolicyReport{
			ObjectMeta: metav1.ObjectMeta{
				Name: "demo-2",
			},
			Summary: appsv1aplha1.PolicyReportSummary{
				Pass: 5,
				Fail: 4,
				Warn: 0,
			},
			Results: []*appsv1aplha1.PolicyReportResult{
				{
					Policy:      "test-policy",
					Rule:        "test-rule",
					Category:    "CIS",
					Result:      "pass",
					Scored:      true,
					Description: "test-description",
					Properties: map[string]string{
						"index":           "1",
						"audit":           "",
						"AuditEnv":        "",
						"AuditConfig":     "",
						"type":            "test-type",
						"remediation":     "test-remediation",
						"test_info":       "test",
						"actual_value":    "test-actual-value",
						"IsMultiple":      "true",
						"expected_result": "test-exp-result",
						"reason":          "test-reason",
					},
				},
			},
		}, "default", "CIS"},
	}
	var kubeconfig *string
	if kubeconfig == nil {
		var path string
		if home := homedir.HomeDir(); home != "" {
			path = filepath.Join(home, ".kube", "config")
		} else {
			path = ""
		}
		kubeconfig = &path
	}

	for _, pr := range policyTests {
		err := createPolicyReport(pr.policyreport.Name, pr.ns, pr.category, kubeconfig, pr.policyreport)
		if err != nil {
			t.Fatalf("error creating policy report: %v", err)
		}
	}
}

func TestRunKubeBench(t *testing.T) {
	_, err := runKubeBench()
	if err != nil {
		t.Fatalf("error getting kube-bench json output due to: %v", err)
	}
}

func TestGetBody(t *testing.T) {
	_, err := getBody()
	if err != nil {
		t.Fatalf("error getting body of kube-bench json output due to: %v", err)
	}
}

func TestGetArguments(t *testing.T) {
	t.Log(getArguments())
}
