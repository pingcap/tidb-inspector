### Usage
```
$ sudo stap cstop_tid.stp <interval>
```

```
$ sudo stap cstop_tid.stp 1
begin to trace top 20 context switch tid
                               Context switch      COUNT
         swapper/18(0)->grafana-server(39107)         34
         grafana-server(39107)->swapper/18(0)         34
              swapper/1(0)->prometheus(33712)         14
              prometheus(33712)->swapper/1(0)         14
          swapper/8(0)->grafana-server(39119)         12
          grafana-server(39119)->swapper/8(0)         12
              swapper/3(0)->prometheus(33741)         12
              prometheus(33741)->swapper/3(0)         12
              swapper/7(0)->prometheus(33712)         12
              prometheus(33712)->swapper/7(0)         12
             swapper/8(0)->tikv-server(37107)         10
             tikv-server(37107)->swapper/8(0)         10
         swapper/9(0)->time-monitor-wo(37117)         10
         time-monitor-wo(37117)->swapper/9(0)         10
         swapper/22(0)->grafana-server(39138)          8
         grafana-server(39138)->swapper/22(0)          8
              swapper/5(0)->prometheus(33712)          8
              prometheus(33712)->swapper/5(0)          8
             swapper/17(0)->prometheus(34728)          7
             prometheus(34728)->swapper/17(0)          7
                                         idle        277
```
For example, `prometheus(33712)` indicates that prometheus thread whose tid is 33712.
