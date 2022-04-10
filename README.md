# NDN HTTP/3 WebTransport Gateway

![NDN-webtrans logo](docs/logo.svg)

This program provides a proxy between Chromium [WebTransport API](https://web.dev/webtransport/) and Named Data Networking's plain UDP transport.
It is designed to work with [NDNts](https://yoursunny.com/p/NDNts/) `@ndn/quic-transport` package.

## Deployment Instructions

1. Install Go compiler.

2. Compile and install this program:

    ```bash
    go install github.com/yoursunny/NDN-webtrans/cmd/ndn-webtrans-gateway@latest
    ```

3. Edit UDP MTU in NFD configuration:

    ```bash
    sudo infoedit -f /etc/ndn/nfd.conf -s face_system.udp.unicast_mtu -v 1200
    ```

4. Obtain an TLS certificate with [acme.sh](https://github.com/acmesh-official/acme.sh).

5. Start this program:

    ```bash
    ndn-webtrans-gateway \
      -cert fullchain.cer -key tls.key \
      -listen :6367 -connect 127.0.0.1:6363
    ```
