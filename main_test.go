// package
package main

import (
	"context"
	"testing"

	appsv1aplha1 "github.com/mritunjaysharma394/policy-report-prototype/pkg/apis/wgpolicyk8s.io/v1alpha1"
	testclient "github.com/mritunjaysharma394/policy-report-prototype/pkg/generated/clientset/versioned/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreatePolicyReport(t *testing.T) {
	policyTests := []struct {
		name         string
		policyreport *appsv1aplha1.PolicyReport
		ns           string
	}{{"demo-test", &appsv1aplha1.PolicyReport{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo",
		},
		Summary: appsv1aplha1.PolicyReportSummary{
			Pass: 10,
			Fail: 4,
			Warn: 0,
		},
	}, "default"},
	}

	for _, pr := range policyTests {
		_, err := testclient.NewSimpleClientset().Wgpolicyk8sV1alpha1().PolicyReports(pr.ns).Create(context.TODO(), pr.policyreport, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("error injecting pod add: %v", err)
		}
	}
}
