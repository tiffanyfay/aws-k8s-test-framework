package framework

import (
	"context"

	"github.com/tiffanyfay/aws-k8s-test-framework/pkg/cloud"
	"github.com/tiffanyfay/aws-k8s-test-framework/test/e2e/framework/resource"
	"github.com/tiffanyfay/aws-k8s-test-framework/test/e2e/framework/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// TODO(@M00nF1sh): migrate to use k8s test framework when it's isolated without pulling in all dependencies.
// This is an simplified version of k8s test framework, since some dependencies don't work under go module :(.
// Framework supports common operations used by e2e tests; it will keep a client & a namespace for you.
type Framework struct {
	ClientSet       clientset.Interface
	Cloud           cloud.Cloud
	ResourceManager *resource.Manager
	Config          *rest.Config // TODO delete me

	Options Options

	// To make sure that this framework cleans up after itself, no matter what,
	// we install a Cleanup action before each test and clear it after.  If we
	// should abort, the AfterSuite hook should run all Cleanup actions.
	cleanupHandle CleanupActionHandle
}

// func NewFastFramework() (*Framework, error) {
// 	var err error

// 	f := &Framework{Options: globalOptions}

// 	f.Config, err = f.buildRestConfig()
// 	if err != nil {
// 		fmt.Println("Failed to make config")
// 		return nil, err
// 	}

// 	f.ClientSet, err = clientset.NewForConfig(f.Config)
// 	if err != nil {
// 		fmt.Println("failed to make clientset")
// 		return nil, err
// 	}

// 	f.AWSClient, err = awsutils.New()
// 	if err != nil {
// 		fmt.Println("failed to make aws client")
// 		return nil, err
// 	}

// 	f.ResourceManager = resource.NewManager(f.ClientSet)

// 	return f, nil
// }

// New makes a new framework and sets up a BeforeEach/AfterEach for you.
func New() *Framework {
	f := &Framework{
		Options: globalOptions,
	}

	BeforeEach(f.BeforeEach)
	AfterEach(f.AfterEach)

	return f
}

// TODO
func (f *Framework) BeforeEach() {
	// The fact that we need this feels like a bug in ginkgo.
	// https://github.com/onsi/ginkgo/issues/222
	if f.ClientSet == nil {
		var err error
		restCfg, err := f.buildRestConfig()
		Expect(err).NotTo(HaveOccurred())
		f.Config = restCfg // TODO delete me
		f.ClientSet, err = clientset.NewForConfig(restCfg)
		Expect(err).NotTo(HaveOccurred())
	}
	if f.Cloud == nil {
		// 	reg := prometheus.NewRegistry()
		// 	mc, _ := metric.NewCollector(reg, "alb")
		var err error
		f.Cloud, err = cloud.New(cloud.Config{
			ClusterName:   f.Options.ClusterName,
			APIMaxRetries: 2,
		})
		Expect(err).NotTo(HaveOccurred())
	}
	f.ResourceManager = resource.NewManager(f.ClientSet)
	f.cleanupHandle = AddCleanupAction(f.cleanupAction())
}

// TODO
func (f *Framework) AfterEach() {
	RemoveCleanupAction(f.cleanupHandle)

	f.cleanupAction()()
}

func (f *Framework) cleanupAction() func() {
	resManager := f.ResourceManager
	return func() {
		if err := resManager.Cleanup(context.TODO()); err != nil {
			utils.Failf("%v", err)
		}
	}
}

func (f *Framework) buildRestConfig() (*rest.Config, error) {
	// restCfg, err := clientcmd.BuildConfigFromFlags("", "")
	restCfg, err := clientcmd.BuildConfigFromFlags("", f.Options.KubeConfig)
	if err != nil {
		return nil, err
	}
	restCfg.QPS = 20
	restCfg.Burst = 50
	return restCfg, nil
}
