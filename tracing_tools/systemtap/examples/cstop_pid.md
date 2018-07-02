### Usage: 
```
$ sudo stap cstop_pid.stp <interval>
```

```
$ sudo stap cstop_pid.stp 1
begin to trace top 20 context switch pid
                               Context switch      COUNT
             swapper/13(0)->prometheus(33709)         25
             prometheus(33709)->swapper/13(0)         25
              swapper/1(0)->prometheus(33709)         13
              prometheus(33709)->swapper/1(0)         13
             swapper/19(0)->prometheus(33709)         11
             prometheus(33709)->swapper/19(0)         11
        swapper/37(0)->time-monitor-wo(36984)         10
        time-monitor-wo(36984)->swapper/37(0)         10
            swapper/10(0)->tikv-server(36984)         10
            tikv-server(36984)->swapper/10(0)         10
             swapper/11(0)->prometheus(33709)          6
             prometheus(33709)->swapper/11(0)          6
             swapper/4(0)->raftstore-1(36984)          6
             raftstore-1(36984)->swapper/4(0)          6
                  swapper/8(0)->stapio(42245)          5
                  stapio(42245)->swapper/8(0)          5
                  swapper/34(0)->rcu_sched(8)          4
                  rcu_sched(8)->swapper/34(0)          4
                   swapper/4(0)->iscsid(2128)          4
                   iscsid(2128)->swapper/4(0)          4
                                         idle        180
```
For example, `prometheus(33709)` indicates prometheus process whose pid is 33709.
