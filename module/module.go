package module

import (
	"os"

	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	pluginapiv1alpha1 "slime.io/slime/modules/plugin/api/v1alpha1"
	"slime.io/slime/modules/plugin/controllers"
	"slime.io/slime/slime-framework/apis/config/v1alpha1"
	istionetworkingapi "slime.io/slime/slime-framework/apis/networking/v1alpha3"
	"slime.io/slime/slime-framework/bootstrap"
	"slime.io/slime/slime-framework/model"
)

const Name = "plugin"

type Module struct {
	config v1alpha1.Plugin
}

func (m *Module) Name() string {
	return Name
}

func (m *Module) Config() proto.Message {
	return &m.config
}

func (m *Module) InitScheme(scheme *runtime.Scheme) error {
	for _, f := range []func(*runtime.Scheme) error{
		clientgoscheme.AddToScheme,
		pluginapiv1alpha1.AddToScheme,
		istionetworkingapi.AddToScheme,
	} {
		if err := f(scheme); err != nil {
			return err
		}
	}
	return nil
}

func (m *Module) InitManager(mgr manager.Manager, env bootstrap.Environment, cbs model.ModuleInitCallbacks) error {
	cfg := &m.config
	if env.Config != nil && env.Config.Plugin != nil {
		cfg = env.Config.Plugin
	}
	_ = cfg // unused until now

	var err error
	if err = (&controllers.PluginManagerReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		log.Errorf("unable to create pluginManager controller, %+v", err)
		os.Exit(1)
	}
	if err = (&controllers.EnvoyPluginReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		log.Errorf("unable to create EnvoyPlugin controller, %+v", err)
		os.Exit(1)
	}

	return nil
}