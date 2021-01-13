# go-work-tcpbalancer
An example of a TCP load balancer using Go concurrency pattern [Job](https://github.com/AgentCoop/go-work).

## Demo
```bash
$ docker-compose up --build
```
Once the client completes its work, you will see resized images of gophers in the ./samples directory. 

## Components
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
  * _--times - run scanning N times
  * _--dry-run_ - dispatches request without image resizing
  * _--debug_ - enables Go profiling
### 💻 Backend server
Handles requests for image resizing.
List of the options:
  * _--port, -p_ - port number to listen to
  * _--name_ - server name
  * _--debug_ - enables Go profiling
  * _--loglevel_ - 0 or 1

### 🌎 Proxy server
Balances incoming requests across a group of backend servers.
List of the options:
  * _--port, -p_ - port number to listen to
  * _-u_ - an upstream server (can by specified multiple times)
  * _-maxconns_ - inbound connections limit
  * _--debug_ - enables Go profiling
  * _--loglevel_ - 0 or 1
