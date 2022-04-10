package main

import (
	"net"
	"net/http"

	"github.com/adriancable/webtransport-go"
	"go.uber.org/zap"
)

func handleGateway(rw http.ResponseWriter, r *http.Request) {
	session := r.Body.(*webtransport.Session)
	logEntry := logger.With(
		zap.String("origin", r.Header.Get("origin")),
		zap.String("client", r.RemoteAddr),
	)

	conn, e := (&net.Dialer{}).DialContext(r.Context(), "udp", *flagRouter)
	if e != nil {
		logEntry.Warn("DialUDP error", zap.Error(e))
		session.RejectSession(504)
		return
	}
	defer conn.Close()
	logEntry = logEntry.With(zap.Stringer("local", conn.LocalAddr()))

	logEntry.Info("accept session")
	session.AcceptSession()
	defer session.CloseSession()

	crPkts, rcPkts := 0, 0
	finish := make(chan error, 1)

	go func() {
		for {
			msg, e := session.ReceiveMessage(session.Context())
			if e != nil {
				finish <- e
				break
			}
			crPkts++
			conn.Write(msg)
		}
	}()

	go func() {
		buf := make([]byte, 9000)
		for {
			n, e := conn.Read(buf)
			if e != nil {
				finish <- e
				break
			}
			rcPkts++
			session.SendMessage(buf[:n])
		}
	}()

	finishReason := <-finish
	logEntry.Info("end session",
		zap.NamedError("reason", finishReason),
		zap.Int("cr-pkts", crPkts),
		zap.Int("rc-pkts", rcPkts),
	)
}
