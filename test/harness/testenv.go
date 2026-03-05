package harness

import (
	"context"
	"path/filepath"
	"runtime"

	appv1alpha1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/k8senv"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/micrologger/microloggertest"
	capi "sigs.k8s.io/cluster-api/api/core/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	capo "github.com/giantswarm/cluster-apps-operator/v3/api/capo/v1alpha4"
	capvcd "github.com/giantswarm/cluster-apps-operator/v3/api/capvcd/v1beta1"
	capz "github.com/giantswarm/cluster-apps-operator/v3/api/capz/v1alpha4"
)

var manager k8senv.Manager

// InitManager creates and initializes the k8senv manager singleton. It must
// be called once before any test acquires an environment (typically from
// TestMain).
func InitManager(ctx context.Context) error {
	_, thisFile, _, _ := runtime.Caller(0)
	crdDir := filepath.Join(filepath.Dir(thisFile), "..", "crds")

	manager = k8senv.NewManager(
		k8senv.WithPoolSize(4),
		k8senv.WithCRDDir(crdDir),
	)

	return manager.Initialize(ctx)
}

// ShutdownManager stops all k8senv instances. Call from TestMain after
// tests complete.
func ShutdownManager() error {
	if manager == nil {
		return nil
	}
	return manager.Shutdown()
}

// TestEnv wraps a single k8senv instance and the Kubernetes clients
// configured with the project's scheme.
type TestEnv struct {
	T          TestingT
	inst       k8senv.Instance
	k8sClient  k8sclient.Interface
	ctrlClient client.Client
	logger     micrologger.Logger
}

// NewTestEnv acquires a k8senv instance, creates Kubernetes clients with
// the same scheme as the production service, and registers cleanup to
// release the instance when the test finishes.
func NewTestEnv(t TestingT) *TestEnv {
	t.Helper()

	ctx := context.Background()

	inst, err := manager.Acquire(ctx)
	if err != nil {
		t.Fatalf("harness: acquire k8senv instance: %v", err)
	}

	cfg, err := inst.Config()
	if err != nil {
		_ = inst.Release()
		t.Fatalf("harness: get rest.Config: %v", err)
	}

	logger := microloggertest.New()

	k8sClients, err := k8sclient.NewClients(k8sclient.ClientsConfig{
		Logger: logger,
		SchemeBuilder: k8sclient.SchemeBuilder{
			appv1alpha1.AddToScheme,
			capi.AddToScheme,
			capo.AddToScheme,
			capz.AddToScheme,
			capvcd.AddToScheme,
		},
		RestConfig: cfg,
	})
	if err != nil {
		_ = inst.Release()
		t.Fatalf("harness: create k8s clients: %v", err)
	}

	env := &TestEnv{
		T:          t,
		inst:       inst,
		k8sClient:  k8sClients,
		ctrlClient: k8sClients.CtrlClient(),
		logger:     logger,
	}

	t.Cleanup(func() {
		if err := inst.Release(); err != nil {
			t.Logf("harness: release instance: %v", err)
		}
	})

	return env
}

// K8sClient returns the k8sclient.Interface for this test environment.
func (e *TestEnv) K8sClient() k8sclient.Interface {
	return e.k8sClient
}

// CtrlClient returns the controller-runtime client for this test environment.
func (e *TestEnv) CtrlClient() client.Client {
	return e.ctrlClient
}

// Logger returns the logger for this test environment.
func (e *TestEnv) Logger() micrologger.Logger {
	return e.logger
}
