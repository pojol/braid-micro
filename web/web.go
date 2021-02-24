package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-echarts/go-echarts/charts"
	"github.com/pojol/braid-go/3rd/consul"
	"github.com/pojol/braid-go/3rd/redis"
	"github.com/pojol/braid-go/modules/linkerredis"
)

var (
	help       bool
	host       string
	tag        string
	redisAddr  string
	consulAddr string
)

func initFlag() {
	flag.BoolVar(&help, "h", false, "this help")

	flag.StringVar(&tag, "tag", "braid", "set discover tag")
	flag.StringVar(&host, "host", ":8888", "set http listen address")
	flag.StringVar(&redisAddr, "redis", "redis://127.0.0.1:6379/0", "set redis address")
	flag.StringVar(&consulAddr, "consul", "http://127.0.0.1:8500", "set consul address")
}

var redisclient *redis.Client

func linkInfo() *charts.Sankey {
	sankey := charts.NewSankey()
	sankey.SetGlobalOptions(charts.TitleOpts{Title: "连接分布图"})

	services, err := consul.GetCatalogServices(consulAddr, tag)
	if err != nil {
		return nil
	}

	// braid_linker-linknum-gateway-base-aad90ab8246a:sharp_rubin:14201
	var sankeyNode []charts.SankeyNode
	var sankeyLink []charts.SankeyLink

	conn := redis.Get().Conn()
	defer conn.Close()

	for _, v := range services {
		parent := v.ServiceName + "-" + v.ServiceID
		sankeyNode = append(sankeyNode, charts.SankeyNode{Name: parent})

		// 从父节点拿到所有的子节点成员
		childs, err := redis.ConnSMembers(conn, linkerredis.LinkerRedisPrefix+"relation-"+v.ServiceName)
		if err != nil {
			continue
		}

		for _, child := range childs {

			nod := strings.Split(child, "-")

			field := linkerredis.LinkerRedisPrefix + "lst-" + v.ServiceName + "-" + nod[0] + "-" + nod[1]
			tokens, err := redis.ConnLRange(conn, field, 0, -1)
			if err != nil || len(tokens) == 0 {
				continue
			}

			sankeyLink = append(sankeyLink, charts.SankeyLink{
				Source: parent,
				Target: nod[0] + "-" + nod[1],
				Value:  float32(len(tokens)),
			})
		}
	}

	sankey.Add("link", sankeyNode, sankeyLink, charts.LabelTextOpts{Show: true})

	return sankey
}

func linkHandler(w http.ResponseWriter, _ *http.Request) {

	connPage := charts.NewPage(charts.RouterOpts{URL: host + "/link", Text: "连接流向图"})
	connPage.InitOpts.PageTitle = "braid-web"
	connPage.Add(
		linkInfo(),
	)

	f, err := os.Create("link.html")
	if err != nil {
		log.Println(err)
	}

	connPage.Render(w, f)

}

func logTracing(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Tracing request for %s\n", r.RequestURI)
		next.ServeHTTP(w, r)
	}
}

func main() {
	initFlag()

	flag.Parse()
	if help {
		flag.Usage()
		return
	}

	redisclient = redis.New()
	redisclient.Init(redis.Config{
		Address:        redisAddr,
		ReadTimeOut:    time.Millisecond * time.Duration(5000),
		WriteTimeOut:   time.Millisecond * time.Duration(5000),
		ConnectTimeOut: time.Millisecond * time.Duration(2000),
		IdleTimeout:    time.Millisecond * time.Duration(0),
		MaxIdle:        16,
		MaxActive:      128,
	})
	defer redisclient.Close()

	http.HandleFunc("/link", logTracing(linkHandler))
	//http.HandleFunc("/stream", logTracing(streamHandler))
	log.Println("Run server at " + host)
	http.ListenAndServe(host, nil)

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	<-ch
}
