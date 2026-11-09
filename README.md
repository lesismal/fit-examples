# fit examples

[![CI](https://github.com/lesismal/fit-examples/actions/workflows/ci.yml/badge.svg)](https://github.com/lesismal/fit-examples/actions/workflows/ci.yml)

[English](README.md) | [简体中文](README.zh-CN.md)

Config examples for [fit](https://github.com/lesismal/fit), the gateway and web server on fib. Each
directory below holds one feature's config: a small one with the fields a user would set, its
comments in English and then in Chinese, in fit's [.conf format](https://github.com/lesismal/fit/blob/main/docs/reference/fit.md), one `fit.conf` file. The same
examples in [YAML](https://github.com/lesismal/fit/blob/main/docs/reference/yaml.md) and [JSON](https://github.com/lesismal/fit/blob/main/docs/reference/json.md)
are planned, see [PLAN.md](PLAN.md). Every field of every module is in fit's
[templates](https://github.com/lesismal/fit/blob/main/templates/README.md) instead. The test runs `fit validate` on every
example, and CI runs it against fit's latest `main` too.

Run an example from this repository's root, since the examples name their files relative to it:

```bash
go install github.com/lesismal/fit/cmd/fit@latest
fit run -c static/fit.conf
```

Addresses such as 10.0.0.11, 127.0.0.1:9000 or auth.example.com stand for your own backends and
services, and host names for your own domains. Passwords are examples: the users of
[basic-auth](basic-auth/fit.conf), [console](console/fit.conf), the LLM feature's and the tunnel's
consoles log in with `example-password`, whose bcrypt hash was made with `fit hash-password`.

## Handlers

| Example | What it shows |
| --- | --- |
| [hello](hello/fit.conf) | A fixed response on port 8080, and 404 for any other path. |
| [static](static/fit.conf) | A file server from [static/www](static/www): directory listings, precompressed twins, small files kept in memory. |
| [proxy](proxy/fit.conf) | A reverse proxy to two upstreams, with limits, timeouts and health checks. |
| [load-balancing](load-balancing/fit.conf) | Weighted least-connections, sticky cookies, active and passive health checks, retries within a budget, a canary split, failover and mirroring. |
| [grpc](grpc/fit.conf) | gRPC over h2c, proxied to a gRPC server over HTTP/2. |
| [fastcgi](fastcgi/fit.conf) | PHP through PHP-FPM on a Unix socket, with a front controller. |
| [forward-proxy](forward-proxy/fit.conf) | An HTTP CONNECT forward proxy, with rules for clients, destinations and ports. |
| [socks](socks/fit.conf) | A SOCKS5 and SOCKS4 forward proxy on a listener of its own. |

## Routing

| Example | What it shows |
| --- | --- |
| [routing](routing/fit.conf) | Routes matched by path, method, header, query, client address, regular expression and expression, with priorities. |
| [virtual-hosts](virtual-hosts/fit.conf) | Sites told apart by host name, exact and wildcard, with a default; `site` blocks in the .conf file. |

## Middleware

| Example | What it shows |
| --- | --- |
| [basic-auth](basic-auth/fit.conf) | HTTP Basic authentication against bcrypt hashes. |
| [jwt](jwt/fit.conf) | JSON Web Tokens verified with a JWK set, issuer and audience, claims passed on as fields. |
| [forward-auth](forward-auth/fit.conf) | Each request asked about at an auth service, its user fields passed on. |
| [client-cert](client-cert/fit.conf) | Client certificates asked for on some routes only, their subject passed on. |
| [rate-limit](rate-limit/fit.conf) | Request rates by client address and by API key, behind trusted proxies. |
| [concurrency-limit](concurrency-limit/fit.conf) | At most so many of a route's requests under way at once. |
| [body-limit](body-limit/fit.conf) | Request body limits per route, beside the listeners' own. |
| [circuit-breaker](circuit-breaker/fit.conf) | A route that stops sending to a failing or slow upstream for a while. |
| [cache](cache/fit.conf) | An HTTP cache by RFC 9111, with a cache lock and stale responses. |
| [compress](compress/fit.conf) | zstd, brotli and gzip compression of responses. |
| [headers](headers/fit.conf) | Security headers set, internal fields removed, request fields set. |
| [rewrite](rewrite/fit.conf) | Paths rewritten by prefix and by regular expression. |
| [redirect](redirect/fit.conf) | Redirects of a host and of a moved section. |
| [wasm](wasm/fit.conf) | A WebAssembly plugin, [continue.wasm](wasm/continue.wasm), from [continue.wat](wasm/continue.wat), on a pool of its own. |

## TLS and listeners

| Example | What it shows |
| --- | --- |
| [https](https/fit.conf) | HTTPS with certificates of your own, and port 80 redirected; the certificate files are not here, so its test does not validate it. |
| [auto-https](auto-https/fit.conf) | HTTPS by default: no listeners written, certificates from Let's Encrypt for the routes' hosts. |
| [acme-on-demand](acme-on-demand/fit.conf) | Certificates obtained during the handshake, for names an ask URL allows, within rate and count limits. |
| [acme-shared-storage](acme-shared-storage/fit.conf) | ACME certificates shared by several instances through Redis. |
| [mtls](mtls/fit.conf) | Mutual TLS: every client must present a certificate of [a CA](mtls/clients-ca.pem). |
| [http3](http3/fit.conf) | HTTP/3 over QUIC beside HTTP/1.1 and HTTP/2 on one port. |
| [proxy-protocol](proxy-protocol/fit.conf) | The PROXY protocol from a trusted load balancer, for the client's real address. |
| [limits](limits/fit.conf) | Every listener's limits, and fib's event loops, workers and buffers. |

## Layer 4

| Example | What it shows |
| --- | --- |
| [l4-tcp](l4-tcp/fit.conf) | TCP passed through to PostgreSQL, and TLS routed by SNI without terminating it, a PROXY header sent upstream. |
| [l4-udp](l4-udp/fit.conf) | UDP datagrams to two DNS resolvers, a session per client address. |

## Providers

| Example | What it shows |
| --- | --- |
| [provider-docker](provider-docker/fit.conf) | Routes from Docker containers' labels. |
| [provider-kubernetes](provider-kubernetes/fit.conf) | The Gateway API and Ingresses of a Kubernetes cluster. |
| [provider-consul](provider-consul/fit.conf) | Routes from the Consul catalog's tags, to healthy instances. |
| [provider-consul-kv](provider-consul-kv/fit.conf) | Config parts from keys in Consul's KV store. |
| [provider-etcd](provider-etcd/fit.conf) | Config parts from keys in etcd. |
| [docker-compose](docker-compose/README.md) | fit's image in front of an app's containers in Compose, routed by their labels. |
| [kubernetes](kubernetes/README.md) | fit's Helm chart in a kind cluster, serving a Gateway and an Ingress. |

## Observability

| Example | What it shows |
| --- | --- |
| [logging](logging/fit.conf) | fit's own log by logger, and the access log. |
| [telemetry](telemetry/fit.conf) | OpenTelemetry traces and metrics over OTLP/HTTP. |
| [metrics](metrics/fit.conf) | Prometheus metrics at the Admin API's /metrics, by listener and route. |

## Admin API and console

| Example | What it shows |
| --- | --- |
| [admin](admin/fit.conf) | The Admin API on a Unix socket, and the config history it keeps. |
| [console](console/fit.conf) | The console's web pages behind a login, and the config kept encrypted in admin.store. |

## LLM and tunnel

| Example | What it shows |
| --- | --- |
| [llm-hub](llm-hub/fit.conf) | The LLM feature's hub, which nodes and proxies link to. |
| [llm-node](llm-node/fit.conf) | A node serving a local Ollama's models through the hub. |
| [llm-proxy](llm-proxy/fit.conf) | A proxy: an OpenAI- and Anthropic-compatible API, for agents, calling the nodes' models. |
| [tunnel-hub](tunnel-hub/fit.conf) | The tunnel's hub, over TCP and WebSocket. |
| [tunnel-node](tunnel-node/fit.conf) | A node behind NAT exposing its services through the hub. |
| [tunnel-proxy](tunnel-proxy/fit.conf) | Local ports forwarded to a node's services. |

The examples of a role pair link to each other on 127.0.0.1, so they run on one machine as they
are, each with `fit run -c` in a terminal of its own. The LLM feature's examples set their `id`, which
fit would otherwise make and write back into the file.

## WebRTC

| Example | What it shows |
| --- | --- |
| [webrtc-turn](webrtc-turn/fit.conf) | A TURN server, with fixed users and time-limited credentials. |
| [whip-whep](whip-whep/fit.conf) | fit as a media server: streams published with WHIP and played with WHEP. |
| [webrtc-proxy](webrtc-proxy/fit.conf) | WHIP and WHEP media servers behind fit, their media relayed through one UDP port. |
| [streaming](streaming/fit.conf) | The streaming role: RTMP, RTSP, SRT and WHIP in, HTTP-FLV, HLS and WHEP out, behind a `media_auth` policy. |
| [media-pages](media-pages/fit.conf) | fit's media pages: a room call, publish and play, with a login and microphone and camera switches. |
| [video-site](video-site/fit.conf) | A video site: films in a media library on demand (`vod`) and as a channel, played with hls.js through playback URLs that expire, which [its own API](video-site/api/main.go) asks `media_tokens` for after logging the viewer in; fit's library and player pages for the staff. [make-media.sh](video-site/make-media.sh) makes its films with ffmpeg. |

## Sync

| Example | What it shows |
| --- | --- |
| [sync-files-to-elasticsearch](sync-files-to-elasticsearch/fit.conf) | Log files tailed across rotations into a daily Elasticsearch index, as Filebeat does, with dead letters. |
| [sync-mysql-to-postgres](sync-mysql-to-postgres/fit.conf) | MySQL's tables kept in PostgreSQL: a snapshot without locks, then the binlog, exactly once. |
| [sync-postgres-to-kafka](sync-postgres-to-kafka/fit.conf) | PostgreSQL's changes through logical replication to Kafka topics in Debezium's format. |
| [sync-rabbitmq-to-kafka](sync-rabbitmq-to-kafka/fit.conf) | A RabbitMQ queue into Kafka, each message acknowledged once Kafka has it, and a Kafka topic published to a RabbitMQ exchange with publisher confirms. |
| [sync-postgres-to-nats](sync-postgres-to-nats/fit.conf) | PostgreSQL's changes to NATS JetStream, deduplicated by Nats-Msg-Id, and the stream read back through a durable pull consumer into a search index. |
| [sync-pulsar-to-elasticsearch](sync-pulsar-to-elasticsearch/fit.conf) | A Pulsar topic into Elasticsearch through readers from the checkpoint's message IDs, and MySQL's changes to Pulsar with broker deduplication. |
| [sync-mqtt-to-influxdb](sync-mqtt-to-influxdb/fit.conf) | The IoT pipeline: devices' JSON readings on MQTT at QoS 1, kept by a persistent session, as InfluxDB points by device and time; alerts back out on MQTT. |
| [sync-mqtt-to-tdengine](sync-mqtt-to-tdengine/fit.conf) | Meters' readings on MQTT into TDengine's super table, a subtable per meter, through taosAdapter's REST interface. |
| [sync-mssql-to-kafka](sync-mssql-to-kafka/fit.conf) | SQL Server's changes from its CDC tables, after a snapshot without locks, to Kafka topics in Debezium's format. |
| [sync-mysql-to-doris](sync-mysql-to-doris/fit.conf) | MySQL's tables kept in Apache Doris (or StarRocks) by Stream Load with labels and the delete sign, and a Doris table polled by a cursor. |
| [sync-files-to-loki](sync-files-to-loki/fit.conf) | Log files to Grafana Loki, streams labelled from bounded fields, and to an OpenTelemetry Collector as OTLP log records. |
| [sync-postgres-to-vectors](sync-postgres-to-vectors/fit.conf) | PostgreSQL's articles with embeddings (the embed transform, cached by the text's hash) upserted into Qdrant and Milvus, deletes deleting. |
| [sync-wasm-transform](sync-wasm-transform/fit.conf) | A WebAssembly transform of your own, [in Go](sync-wasm-transform/plugin/main.go), over records read from a file; runs with fit alone. |
| [sync-wasm-codec](sync-wasm-codec/fit.conf) | A WebAssembly codec of your own, [in Go](sync-wasm-codec/plugin/main.go), reading a line format of its own from a file and writing it back; runs with fit alone. |
| [sync-wasm-source](sync-wasm-source/fit.conf) | A WebAssembly source of your own, [in Go](sync-wasm-source/plugin/main.go), reading a paged HTTP API, which fit serves too, from its checkpoint. |
| [sync-wasm-sink](sync-wasm-sink/fit.conf) | A WebAssembly sink of your own, [in Go](sync-wasm-sink/plugin/main.go), posting batches to a webhook, which fit answers too. |

The WebAssembly plugins build by Go or by TinyGo, as each plugin's comment shows. The sync examples
read their passwords from environment variables, which `fit validate` reads too.
[Sync](https://github.com/lesismal/fit/blob/main/docs/sync.md) explains pipelines, checkpoints and
every connector.

## BitTorrent

| Example | What it shows |
| --- | --- |
| [bittorrent](bittorrent/fit.conf) | The BitTorrent client, with a watch directory, categories (one feeding a media library), a queue, weekday alternative rates and [bittorrent/releases](bittorrent/releases) published as hybrid torrents with fit's own web seed; a private tracker on passkeys of a `media_auth` policy; files played at `/bt/<info hash>/` while they download; `fit bt` through the Admin API. |

[BitTorrent](https://github.com/lesismal/fit/blob/main/docs/bittorrent.md) explains the client, the
tracker, publishing, `bt_stream` and `fit bt`.

## Modules

| Example | What it shows |
| --- | --- |
| [module/hello](module/hello/hello.go) | A middleware of your own, built into fit with `fit build`. |

The `test.sh` of [docker-compose](docker-compose/test.sh) and [kubernetes](kubernetes/test.sh)
runs its example as its README says and checks the answers; CI runs them. How fit is packaged is
in [deploy](https://github.com/lesismal/fit/blob/main/deploy/README.md).
