module github.com/AgentCoop/go-work-tcpbalancer

go 1.15

//replace github.com/AgentCoop/go-work => /home/pihpah/go/src/github.com/AgentCoop/go-work
//replace github.com/AgentCoop/net-dataframe => /home/pihpah/go/src/github.com/AgentCoop/net-dataframe
//replace github.com/AgentCoop/net-manager => /home/pihpah/go/src/github.com/AgentCoop/net-manager

require (
	github.com/AgentCoop/go-work v0.0.1
	github.com/AgentCoop/net-manager v0.0.0-20210114144114-c04e8b885085
	github.com/jessevdk/go-flags v1.4.0
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
)
