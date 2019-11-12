# consul-service-has-ip

API server for checking the ip is included in a service.

It's able to know the IP is alive easier than use consul api directory.

## How to

run api server

```
$ ./consul-service-has-ip --consul-api-endpoint http://consul.service.xx.consul:8500
```

checking ip is in a service.

```
curl -v localhost:3000/has/api-xx/10.0.20.x
* About to connect() to localhost port 3000 (#0)
*   Trying 127.0.0.1...
* Connected to localhost (127.0.0.1) port 3000 (#0)
> GET /has/api-xx/10.0.20.x HTTP/1.1
> User-Agent: curl/7.29.0
> Host: localhost:3000
> Accept: */*
>
< HTTP/1.1 404 Not Found
< Date: Tue, 12 Nov 2019 04:28:54 GMT
< Content-Length: 71
< Content-Type: text/plain; charset=utf-8
<
{"error":1,"messages":"ip:10.0.20.x is not in service:api-xx"}

* Connection #0 to host localhost left intact
[kazeburo@manage2 ~]$ curl -v localhost:3000/has/api-xx/10.0.9.x
* About to connect() to localhost port 3000 (#0)
*   Trying 127.0.0.1...
* Connected to localhost (127.0.0.1) port 3000 (#0)
> GET /has/api-xx/10.0.9.x HTTP/1.1
> User-Agent: curl/7.29.0
> Host: localhost:3000
> Accept: */*
>
< HTTP/1.1 200 OK
< Date: Tue, 12 Nov 2019 04:29:03 GMT
< Content-Length: 12
< Content-Type: text/plain; charset=utf-8
<
{"error":0}
```

## Usage

```
% ./consul-service-has-ip -h
Usage:
  consul-service-has-ip [OPTIONS]

Application Options:
  -l, --listen=              address to bind (default: 0.0.0.0)
  -p, --port=                Port number to bind (default: 3000)
  -v, --version              Show version
      --read-timeout=        timeout of reading request (default: 30s)
      --write-timeout=       timeout of writing response (default: 90s)
      --shutdown-timeout=    Timeout to wait for all connections to be closed (default: 10s)
      --timeout=             timeout to reques to consul (default: 30s)
      --access-log-dir=      directory to store logfiles
      --access-log-rotate=   Number of day before remove logs (default: 30)
      --consul-api-endpoint= api endpoint of consul. required

Help Options:
  -h, --help                 Show this help message

```
