package report

import (
	"context"
	"fmt"
	"log"

	policyreport "github.com/mritunjaysharma394/policy-report-prototype/pkg/apis/wgpolicyk8s.io/v1alpha2"

	client "github.com/mritunjaysharma394/policy-report-prototype/pkg/generated/v1alpha2/clientset/versioned"
	"k8s.io/apimachinery/pkg/api/errors"
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

	policyReport := clientset.Wgpolicyk8sV1alpha2().PolicyReports(namespace)

	// Check for existing Policy Reports
	result, getErr := policyReport.Get(context.TODO(), r.Name, metav1.GetOptions{})
	// Create new Policy Report if not found
	if errors.IsNotFound(getErr) {
		fmt.Println("creating policy report...")

		result, err = policyReport.Create(context.TODO(), r, metav1.CreateOptions{})
		if err != nil {
			return nil, err
		}
	} else {

		// Update existing Policy Report
		fmt.Println("updating policy report...")
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {

			getObj, err := policyReport.Get(context.TODO(), r.GetName(), metav1.GetOptions{})
			if errors.IsNotFound(err) {
				// This doesnt ever happen even if it is already deleted or not found
				log.Printf("%v not found", r.GetName())
				return nil
			}

			if err != nil {
				return err
			}

			r.SetResourceVersion(getObj.GetResourceVersion())

			_, updateErr := policyReport.Update(context.TODO(), r, metav1.UpdateOptions{})
			return updateErr
		})
		if retryErr != nil {
			panic(fmt.Errorf("update failed: %v", retryErr))
		}
		fmt.Println("updated policy report...")
	}

	return result, nil
}
