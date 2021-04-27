package electork8s

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/elector"
	"github.com/pojol/braid-go/module/logger"
	"github.com/pojol/braid-go/module/mailbox"
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

func newK8sElector() module.Builder {
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

func (*k8sElectorBuilder) Type() string {
	return module.TyElector
}

func (eb *k8sElectorBuilder) AddOption(opt interface{}) {
	eb.opts = append(eb.opts, opt)
}

func (eb *k8sElectorBuilder) Build(serviceName string, mb mailbox.IMailbox, logger logger.ILogger) (module.IModule, error) {

	p := Parm{
		ServiceName: serviceName,
		Namespace:   "default",
		RetryPeriod: time.Second * 2,
	}
	for _, opt := range eb.opts {
		opt.(Option)(&p)
	}

	el := &k8sElector{
		parm:   p,
		mb:     mb,
		logger: logger,
	}

	return el, nil
}

func (e *k8sElector) Init() error {
	clientset, err := newClientset(e.parm.KubeCfg)
	if err != nil {
		return fmt.Errorf("%v Dependency check error %v [%v]", e.parm.ServiceName, "k8s", e.parm.KubeCfg)
	}

	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      "braid-lock",
			Namespace: e.parm.Namespace,
		},
		Client: clientset.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: e.parm.NodID,
		},
	}

	elector, err := leaderelection.NewLeaderElector(leaderelection.LeaderElectionConfig{
		Lock:            lock,
		ReleaseOnCancel: true,
		LeaseDuration:   30 * time.Second,   // 租约时间
		RenewDeadline:   10 * time.Second,   // 更新租约时间
		RetryPeriod:     e.parm.RetryPeriod, // 非master节点的重试时间
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {},
			OnStoppedLeading: func() {},
			OnNewLeader: func(identity string) {
				if identity == e.parm.NodID {
					e.mb.Pub(mailbox.Proc, elector.ElectorStateChange, elector.EncodeStateChangeMsg(elector.EMaster))
					e.logger.Debugf("new leader %s %s", e.parm.NodID, identity)

				} else {
					e.mb.Pub(mailbox.Proc, elector.ElectorStateChange, elector.EncodeStateChangeMsg(elector.ESlave))
				}
			},
		},
	})
	if err != nil {
		return fmt.Errorf("%v Dependency check error %v [%v]", e.parm.ServiceName, "k8s", err.Error())
	}

	e.elector = elector
	return nil
}

type k8sElector struct {
	parm    Parm
	logger  logger.ILogger
	mb      mailbox.IMailbox
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
	module.Register(newK8sElector())
}
