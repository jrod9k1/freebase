# Freebase
A sensible albeit primitive A2S_INFO cache service written in Go. Proof of concept for A2S experimentation in gmod / source engine games.

### Usage

This agent is intended to be run on a BGP Anycast network directly on the edge routers. Iptables is responsible for filtering out A2S_INFO packets bound for the source server then DNAT redirecting them to the bind port on this agent.

See this rule as an example of a correct DNAT/REDIRECT setup:

`iptables -t nat -I PREROUTING -p udp -d 142.202.137.10 --dport 27015 -m string --algo bm --hex-string '|ffffffff54536f7572636520456e67696e6520517565727900|' -j REDIRECT --to-port 27016`

This has been tested running directly on Quagga edge routers out of AS137909.

### Arguments

-ip: Anycast IP of source server you are caching for\
-port: Port of aforementioned server\
-bport: "Bind" port, whatever port you want the agent to serve DNAT'd requests on\
-refresh: How often in seconds the agent will request fresh A2S_INFO data from the source\

`./freebase -ip 142.202.137.10 -port 27015 -bport 27016 -refresh 30`

### Read More About It

![](https://i.jrod.sh/2020/01/firefox_A02KUuDS52.png)
