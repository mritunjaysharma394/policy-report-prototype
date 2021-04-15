// package
package main

import (
	"flag"
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
		}, "default"},
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
					Category:    "test-category",
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
		}, "default"},
	}

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	for _, pr := range policyTests {
		err := createPolicyReport(pr.ns, pr.policyreport, kubeconfig)
		if err != nil {
			t.Fatalf("error creating policy report: %v", err)
		}
	}
}
