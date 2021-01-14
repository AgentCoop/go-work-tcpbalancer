# go-work-tcpbalancer
An example of a TCP load balancer using Go concurrency pattern [Job](https://github.com/AgentCoop/go-work).
<img align="center" style="margin: 6px" src="https://raw.githubusercontent.com/AgentCoop/go-work-tcpbalancer/master/assets/balancer-schema.png" alt='Balancer Schema' aria-label='' />

## Demo
```bash
$ docker-compose up --build
```
When started the client will download few images of Go gophers and send them to the proxy server for resizing.
Once the client completes its work, you will see the resized images of gophers in the _./samples_ directory. 

## Details on implementation
The central part of the balancer implementation is [Job component](https://github.com/AgentCoop/go-work), so let's see
how it's being used in the applications.
  1. [Client](./docs/client.md)
  2. [Proxy Server](./docs/proxy.md)
  3. [Backend Server](./docs/backend.md)
## Applications
### 🚹 Client
The client application scans specified directory for images and sends them to the proxy server for resizing. 
List of the options:
  * _--proxy_ - address of the proxy server to connect to
  * _--loglevel_ - 0 or 1
  * _--minconns_ - minimum number of concurrent connections
  * _--maxconns_ - maximum number of concurrent connections
  * _--input_ - input directory to scan
  * _--output_ - output directory for the resized images
  * _-w, -h_ - target image width and height
  * _--times_ - run scanning N times
  * _--dry-run_ - dispatches request without image resizing
  * _--debug_ - enables Go profiling
### 💻 Backend Server
Handles requests for image resizing.
List of the options:
  * _--port, -p_ - port number to listen to
  * _--name_ - server name
  * _--debug_ - enables Go profiling
  * _--loglevel_ - 0 or 1

### 🌎 Proxy Server
Balances incoming requests across a group of backend servers.
List of the options:
  * _--port, -p_ - port number to listen to
  * _-u_ - an upstream server (can by specified multiple times)
  * _-maxconns_ - inbound connections limit
  * _--debug_ - enables Go profiling
  * _--loglevel_ - 0 or 1
