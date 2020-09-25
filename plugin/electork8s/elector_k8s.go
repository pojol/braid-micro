package electork8s

import (
	"context"
	"errors"
	"time"

	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/module/elector"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

const (
	// ElectionName 基于kubernetes接口实现的选举器
	ElectionName = "K8sElector"
)

var (
	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("convert config error")
)

type k8sElectorBuilder struct {
	opts []interface{}
}

func newK8sElector() elector.Builder {
	return &k8sElectorBuilder{}
}

func onStartedLeading(ctx context.Context) {

}

func onStoppedLeading() {

}

func getConfig(cfg string) (*rest.Config, error) {
	if cfg == "" {
		return rest.InClusterConfig()
	}
	return clientcmd.BuildConfigFromFlags("", cfg)
}

func newClientset(filename string) (*kubernetes.Clientset, error) {
	config, err := getConfig(filename)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func (*k8sElectorBuilder) Name() string {
	return ElectionName
}

func (eb *k8sElectorBuilder) AddOption(opt interface{}) {
	eb.opts = append(eb.opts, opt)
}

func (eb *k8sElectorBuilder) Build(serviceName string) (elector.IElection, error) {

	p := Parm{
		ServiceName: serviceName,
		Namespace:   "default",
		RetryPeriod: time.Second * 2,
	}

	clientset, err := newClientset(p.KubeCfg)
	if err != nil {
		return nil, err
	}

	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      "braid-lock",
			Namespace: p.Namespace,
		},
		Client: clientset.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: p.NodID,
		},
	}

	elector, err := leaderelection.NewLeaderElector(leaderelection.LeaderElectionConfig{
		Lock:            lock,
		ReleaseOnCancel: true,
		LeaseDuration:   30 * time.Second, // 租约时间
		RenewDeadline:   10 * time.Second, // 更新租约时间
		RetryPeriod:     p.RetryPeriod,    // 非master节点的重试时间
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {},
			OnStoppedLeading: func() {},
			OnNewLeader: func(identity string) {
				if identity == p.NodID {
					log.SysElection(p.NodID, identity)
				}
			},
		},
	})
	if err != nil {
		return nil, err
	}

	el := &k8sElector{
		parm:    p,
		elector: elector,
	}

	return el, nil
}

type k8sElector struct {
	parm    Parm
	lock    *resourcelock.LeaseLock
	elector *leaderelection.LeaderElector
}

func (e *k8sElector) IsMaster() bool {
	return e.elector.IsLeader()
}

func (e *k8sElector) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	e.elector.Run(ctx)
}

func (e *k8sElector) Close() {

}

func init() {
	elector.Register(newK8sElector())
}
