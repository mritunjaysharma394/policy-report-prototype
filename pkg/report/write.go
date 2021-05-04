package report

import (
	"context"
	"fmt"

	// TODO: upgrade to v1alpha2 CRD and create a Makefile / shell script for code generation
	// see example at: https://leftasexercise.com/2019/07/29/building-a-bitcoin-controller-for-kubernetes-part-ii/
	// https://github.com/christianb93/bitcoin-controller/blob/master/build/controller/generate_code.sh
	policyreport "github.com/mritunjaysharma394/policy-report-prototype/pkg/apis/wgpolicyk8s.io/v1alpha1"

	client "github.com/mritunjaysharma394/policy-report-prototype/pkg/generated/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
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

	result, err := policyReport.Create(context.TODO(), r, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	// Update Policy Report
	fmt.Println("Updating deployment...")
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Policy Report before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		result, getErr := policyReport.Get(context.TODO(), r.Name, metav1.GetOptions{})
		if getErr != nil {
			panic(fmt.Errorf("failed to get latest version of Policy Report: %v", getErr))
		}

		_, updateErr := policyReport.Update(context.TODO(), result, metav1.UpdateOptions{})
		return updateErr
	})
	if retryErr != nil {
		panic(fmt.Errorf("update failed: %v", retryErr))
	}
	fmt.Println("Updated Policy Report...")

	return result, nil
}
