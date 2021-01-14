# go-work-tcpbalancer
An example of a TCP load balancer using Go concurrency pattern [Job](https://github.com/AgentCoop/go-work).
<img align="center" style="margin: 6px" src="https://raw.githubusercontent.com/AgentCoop/go-work-tcpbalancer/master/assets/balancer-schema.png" alt='Balancer Schema' aria-label='' />

## Demo
```bash
$ docker-compose up --build
```
When started the client will download few images of Go gophers and send them to the proxy server for resizing.
Once the client completes its work, you will see the resized images of gophers in the _./samples_ directory. 

## Benchmarking
<details>
  <summary>Machine</summary>
<pre>
Architecture:                    x86_64
CPU op-mode(s):                  32-bit, 64-bit
Byte Order:                      Little Endian
Address sizes:                   39 bits physical, 48 bits virtual
CPU(s):                          4
On-line CPU(s) list:             0-3
Thread(s) per core:              2
Core(s) per socket:              2
Socket(s):                       1
NUMA node(s):                    1
Vendor ID:                       GenuineIntel
CPU family:                      6
Model:                           78
Model name:                      Intel(R) Core(TM) i5-6200U CPU @ 2.30GHz
Stepping:                        3
CPU MHz:                         2700.012
CPU max MHz:                     2800.0000
CPU min MHz:                     400.0000
BogoMIPS:                        4801.00
Virtualization:                  VT-x
L1d cache:                       64 KiB
L1i cache:                       64 KiB
L2 cache:                        512 KiB
L3 cache:                        3 MiB
NUMA node0 CPU(s):               0-3
...
Flags:                           fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush dts acpi mmx fxsr sse sse2 ss ht tm pbe syscall nx pdpe1gb rdtscp lm constant_tsc art arch_perfmon p
                                 ebs bts rep_good nopl xtopology nonstop_tsc cpuid aperfmperf pni pclmulqdq dtes64 monitor ds_cpl vmx est tm2 ssse3 sdbg fma cx16 xtpr pdcm pcid sse4_1 sse4_2 x2apic movbe popcnt ts
                                 c_deadline_timer aes xsave avx f16c rdrand lahf_lm abm 3dnowprefetch cpuid_fault epb invpcid_single pti ssbd ibrs ibpb stibp tpr_shadow vnmi flexpriority ept vpid ept_ad fsgsbase t
                                 sc_adjust bmi1 avx2 smep bmi2 erms invpcid mpx rdseed adx smap clflushopt intel_pt xsaveopt xsavec xgetbv1 xsaves dtherm ida arat pln pts hwp hwp_notify hwp_act_window hwp_epp md_c
                                 lear flush_l1d
</pre>
</details>

### Floor it
Direct client to backend server with 100 concurrent connections, dry run without the actual image resizing:
```bash
$ go run cmd/backend/*.go -p 9092 --name=localhost --loglevel=0
$ go run cmd/frontend/*.go --proxy localhost:9092 --input=./images --output=./resized -w 200 -h 200 --times=100 --minconns=100 --loglevel=0 --dry-run
-- [ Network Statistics ] --
	bytes sent: 6.4400 Mb
	bytes received: 7.0100 Mb
	Requests Per Second: 1076.02
```

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
