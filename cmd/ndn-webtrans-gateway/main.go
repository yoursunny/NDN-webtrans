// Command ndn-webtrans-gateway accepts HTTP/3 WebTransport datagrams and forwards them to a UDP server.
package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/adriancable/webtransport-go"
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

func main() {
	flag.Parse()

	http.HandleFunc("/ndn", handleGateway)

	server := &webtransport.Server{
		ListenAddr: *flagListen,
		TLSCert:    webtransport.CertFile{Path: *flagCert},
		TLSKey:     webtransport.CertFile{Path: *flagKey},
		QuicConfig: &webtransport.QuicConfig{
			MaxIdleTimeout:          60 * time.Second,
			KeepAlive:               true,
			DisablePathMTUDiscovery: true,
		},
	}

	if e := server.Run(context.Background()); e != nil {
		logger.Fatal("server.Run error", zap.Error(e))
	}
}
