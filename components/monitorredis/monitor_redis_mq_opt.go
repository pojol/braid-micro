package monitorredis

type MqWatchParm struct {
	Prot string
}

type MqWatchOption func(*MqWatchParm)

func WithWatchProt(prot string) MqWatchOption {
	return func(parm *MqWatchParm) {
		parm.Prot = prot
	}
}
