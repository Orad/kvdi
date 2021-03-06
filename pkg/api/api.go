package api

import (
	"net/http"
	"time"

	"github.com/tinyzimmer/kvdi/pkg/apis"
	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/auth"
	"github.com/tinyzimmer/kvdi/pkg/auth/mfa"
	"github.com/tinyzimmer/kvdi/pkg/secrets"
	"github.com/tinyzimmer/kvdi/pkg/util/k8sutil"

	"github.com/gorilla/mux"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var apiLogger = logf.Log.WithName("api")

// DesktopAPI serves HTTP requests for the /api resource
type DesktopAPI interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}

// desktopAPI implements the DesktopAPI interface
type desktopAPI struct {
	// easy for quick read/write operators
	client client.Client
	// config for building rest clients if needed
	restConfig *rest.Config
	// scheme for building rest clients
	scheme *runtime.Scheme
	// the router interface
	router *mux.Router
	// our parent vdi cluster
	vdiCluster *v1alpha1.VDICluster
	// the user auth provider
	auth v1alpha1.AuthProvider
	// the secrets backend
	secrets *secrets.SecretEngine
	// the mfa backend for setting and retrieving OTP secrets
	mfa *mfa.Manager
}

// NewFromConfig builds a new API router from the given kubernetes client configuration
// and vdi cluster name.
func NewFromConfig(cfg *rest.Config, vdiCluster string) (DesktopAPI, error) {
	// build our scheme
	scheme := runtime.NewScheme()
	if err := apis.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	// build a client
	client, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, err
	}

	// retrieve the vdicluster we are working for
	apiLogger.Info("Retrieving VDICluster configuration")
	var found *v1alpha1.VDICluster
	for found == nil {
		if found, err = k8sutil.LookupClusterByName(client, vdiCluster); err != nil {
			apiLogger.Error(err, "Failed to retrieve VDICluster configuration, retrying in 2 seconds...")
			found = nil
			time.Sleep(time.Duration(2) * time.Second)
		}
	}

	// setup the secrets engine
	secretsEngine := secrets.GetSecretEngine(found)
	if err := secretsEngine.Setup(client, found); err != nil {
		return nil, err
	}

	// setup the auth provider
	authProvider := auth.GetAuthProvider(found)
	if err := authProvider.Setup(client, found); err != nil {
		return nil, err
	}

	api := &desktopAPI{
		client:     client,
		restConfig: cfg,
		scheme:     scheme,
		vdiCluster: found,
		auth:       authProvider,
		secrets:    secretsEngine,
		mfa:        mfa.NewManager(secretsEngine),
	}

	return api, api.buildRouter()
}
