package app

import (
	"strconv"
	"testing"

	"github.com/giantswarm/apiextensions-application/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_hasAppChanged(t *testing.T) {
	testCases := []struct {
		name    string
		current *v1alpha1.App
		desired *v1alpha1.App
		result  bool
	}{
		{
			name: "return false when inputs match",
			current: &v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "eggs2-app-operator",
					Namespace: "org-test",
					Labels: map[string]string{
						"giantswarm.io/cluster": "eggs2",
					},
				},
				Spec: v1alpha1.AppSpec{
					Catalog: "control-plane-catalog",
					Name:    "app-operator",
				},
			},
			desired: &v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "eggs2-app-operator",
					Namespace: "org-test",
					Labels: map[string]string{
						"giantswarm.io/cluster": "eggs2",
					},
				},
				Spec: v1alpha1.AppSpec{
					Catalog: "control-plane-catalog",
					Name:    "app-operator",
				},
			},
			result: false,
		},
		{
			name: "return true when spec does not match",
			current: &v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "eggs2-app-operator",
					Namespace: "org-test",
					Labels: map[string]string{
						"giantswarm.io/cluster": "eggs2",
					},
				},
				Spec: v1alpha1.AppSpec{
					Catalog: "control-plane-catalog",
					Name:    "app-operator",
				},
			},
			desired: &v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "eggs2-app-operator",
					Namespace: "org-test",
					Labels: map[string]string{
						"giantswarm.io/cluster": "eggs2",
					},
				},
				Spec: v1alpha1.AppSpec{
					Catalog: "control-plane-test-catalog",
					Name:    "app-operator",
				},
			},
			result: true,
		},
		{
			name: "return true when labels do not match",
			current: &v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "eggs2-app-operator",
					Namespace: "org-test",
					Labels: map[string]string{
						"giantswarm.io/cluster": "eggs2",
					},
				},
				Spec: v1alpha1.AppSpec{
					Catalog: "control-plane-catalog",
					Name:    "app-operator",
				},
			},
			desired: &v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "eggs2-app-operator",
					Namespace: "org-test",
					Labels: map[string]string{
						"giantswarm.io/cluster": "eggs2",
						"test":                  "label",
					},
				},
				Spec: v1alpha1.AppSpec{
					Catalog: "control-plane-test-catalog",
					Name:    "app-operator",
				},
			},
			result: true,
		},
		{
			name: "return true when annotations do not match",
			current: &v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "eggs2-app-operator",
					Namespace: "org-test",
				},
				Spec: v1alpha1.AppSpec{
					Catalog: "control-plane-catalog",
					Name:    "app-operator",
				},
			},
			desired: &v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "eggs2-app-operator",
					Namespace: "org-test",
					Annotations: map[string]string{
						"app-operator.giantswarm.io/latest-configmap-version": "1",
					},
				},
				Spec: v1alpha1.AppSpec{
					Catalog: "control-plane-test-catalog",
					Name:    "app-operator",
				},
			},
			result: true,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			if tc.result != hasAppChanged(tc.current, tc.desired) {
				t.Fatalf("expected %t got %t", tc.result, hasAppChanged(tc.current, tc.desired))
			}
		})
	}
}
