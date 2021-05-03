package report

import (
	"context"

	// TODO: upgrade to v1alpha2 CRD and create a Makefile / shell script for code generation
	// see example at: https://leftasexercise.com/2019/07/29/building-a-bitcoin-controller-for-kubernetes-part-ii/
	// https://github.com/christianb93/bitcoin-controller/blob/master/build/controller/generate_code.sh
	policyreport "github.com/mritunjaysharma394/policy-report-prototype/pkg/apis/wgpolicyk8s.io/v1alpha1"

	client "github.com/mritunjaysharma394/policy-report-prototype/pkg/generated/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

func Write(r *policyreport.PolicyReport, namespace string, kubeconfig string) (*policyreport.PolicyReport, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := client.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	policyReport := clientset.Wgpolicyk8sV1alpha1().PolicyReports(namespace)

	// TODO: check for existing report. If a report exists with the same name, we need to update it.
	result, err := policyReport.Create(context.TODO(), r, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return result, nil
}
