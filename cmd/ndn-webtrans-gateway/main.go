// Command ndn-webtrans-gateway accepts HTTP/3 WebTransport datagrams and forwards them to a UDP server.
package main

import (
	"crypto/tls"
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

var wtServer *webtransport.Server

func main() {
	flag.Parse()

	cert, e := tls.LoadX509KeyPair(*flagCert, *flagKey)
	if e != nil {
		logger.Fatal("tls.LoadX509KeyPair error", zap.Error(e))
	}
	logger.Info("certificate loaded",
		zap.Strings("dns-names", cert.Leaf.DNSNames),
		zap.Time("not-before", cert.Leaf.NotBefore),
		zap.Time("not-after", cert.Leaf.NotAfter),
	)

	h3Server := &http3.Server{
		Addr: *flagListen,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			NextProtos:   []string{http3.NextProtoH3},
		},
		QUICConfig: &quic.Config{
			MaxIdleTimeout:          60 * time.Second,
			KeepAlivePeriod:         30 * time.Second,
			DisablePathMTUDiscovery: true,
		},
	}
	webtransport.ConfigureHTTP3Server(h3Server)
	wtServer = &webtransport.Server{
		H3: h3Server,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	http.HandleFunc("/ndn", handleGateway)

	if e := wtServer.ListenAndServe(); e != nil {
		logger.Fatal("server.Run error", zap.Error(e))
	}
}
