// exporter.go - define the radosgw_exporter executable program

package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const keepAlivePeriod = 10 * time.Minute

var (
	listenAddr  = flag.String("addr", "127.0.0.1:9129", "listen address for radosgw exporter")
	metricsPath = flag.String("path", "/metrics", "URL path for collecting radosgw metrics")
	adminAK     = flag.String("ak", "", "access key id of the admin user of radosgw service")
	adminSK     = flag.String("sk", "", "secret access key of the admin user of radosgw service")
	endpoint    = flag.String("endpoint", "127.0.0.1:8080", "endpoint of the radosgw service")
)

func main() {
	flag.Parse()

	// Check the arguments
	if len(*adminAK) == 0 || len(*adminSK) == 0 {
		fmt.Printf("invalid admin AK/SK for the radosgw service")
		return
	}
	tcpAddr, err := net.ResolveTCPAddr("tcp", *listenAddr)
	if err != nil {
		fmt.Printf("invalid listen address of TCP: %v", err)
		return
	}

	// Register the handlers
	collector, err := NewRadosgwCollector(*endpoint, *adminAK, *adminSK)
	if err != nil {
		fmt.Printf("create radosgw collector failed: %v", err)
		return
	}
	err = prometheus.Register(collector)
	if err != nil {
		fmt.Printf("register the collector failed: %v", err)
		return
	}
	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(200)
		w.Write([]byte(`
            <html>
                <head><title>Radosgw Exporter</title></head>
                <body>
                    <h1>Radosgw Exporter</h1>
                    <p><a href='` + *metricsPath + `'>Metrics</a></p>
                </body>
            </html>`))
	})

	// Listen and serve
	tcpListener, err := net.ListenTCP("tcp", tcpAddr)
	err = http.Serve(customTCPListener{tcpListener}, nil)
	if err != nil {
		fmt.Printf("radosgw exporter occurs error: %v", err)
	}
}

type customTCPListener struct {
	*net.TCPListener
}

func (l customTCPListener) Accept() (net.Conn, error) {
	tc, err := l.AcceptTCP()
	if err != nil {
		return nil, err
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(keepAlivePeriod)
	tc.SetNoDelay(true)
	tc.SetLinger(0)
	return tc, nil
}
