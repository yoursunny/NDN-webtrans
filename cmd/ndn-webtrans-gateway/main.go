// Command ndn-webtrans-gateway accepts HTTP/3 WebTransport datagrams and forwards them to a UDP server.
package main

import (
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/webtransport-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger = func() *zap.Logger {
	var lvl zapcore.Level
	if environ, ok := os.LookupEnv("NDN_WEBTRANS_LOG"); ok {
		lvl.Set(environ)
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		os.Stderr,
		lvl,
	)
	return zap.New(core)
}()

var (
	flagCert   = flag.String("cert", "", "TLS certificate file")
	flagKey    = flag.String("key", "", "TLS key file")
	flagListen = flag.String("listen", "127.0.0.1:6367", "HTTP/3 server address and port")
	flagRouter = flag.String("router", "127.0.0.1:6363", "router address and port")
)

var server *webtransport.Server

func main() {
	flag.Parse()

	http.HandleFunc("/ndn", handleGateway)

	server = &webtransport.Server{
		H3: http3.Server{
			Addr: *flagListen,
			QUICConfig: &quic.Config{
				MaxIdleTimeout:          60 * time.Second,
				KeepAlivePeriod:         30 * time.Second,
				DisablePathMTUDiscovery: true,
			},
		},
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	if e := server.ListenAndServeTLS(*flagCert, *flagKey); e != nil {
		logger.Fatal("server.Run error", zap.Error(e))
	}
}
