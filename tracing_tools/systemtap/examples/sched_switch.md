### Usage
```
sudo stap sched_switch.stp -x <pid>
```

> This script works similar to ftrace's sched_switch. It displays a list of
processes which get switched in and out of the scheduler. The format of display
is PROCESS_NAME PROCESS_PID CPU TIMESTAMP PID: PRIORITY: PROCESS STATE ->/+
NEXT_PID : NEXT_PRIORITY: NEXT_STATE NEXT_PROCESS_NAME
==> indicates that prev process is scheduled out and the next process is
scheduled in.
\+ indicates that prev process has woken up the next process.

```
$ sudo stap sched_switch.stp -x 29404
PROCESS_NAME      CPU           TIMESTAMP   PID: PRIO:S  NEXT_PID: PRIO:S NEXT_PROCESS_NAME
swapper/1           1 1513061754289956125     0:   20:R +   29404:   20: tikv-server
swapper/1           1 1513061754289970469     0:  120:R ==> 29404:  120:R tikv-server
tikv-server         1 1513061754289989216 29404:  120:S ==>     0:  120:R swapper/1
swapper/1           1 1513061754337955756     0:   20:R +   29404:   20: tikv-server
swapper/1           1 1513061754337970086     0:  120:R ==> 29404:  120:R tikv-server
tikv-server         1 1513061754337989086 29404:  120:S ==>     0:  120:R swapper/1
swapper/31         31 1513061754366201334     0:   20:R +   29404:   20: time-monitor-wo
swapper/31         31 1513061754366216764     0:  120:R ==> 29404:  120:R time-monitor-wo
time-monitor-wo    31 1513061754366253877 29404:  120:S ==>     0:  120:R swapper/31
swapper/1           1 1513061754373957481     0:   20:R +   29404:   20: tikv-server
swapper/1           1 1513061754373972888     0:  120:R ==> 29404:  120:R tikv-server
tikv-server         1 1513061754373993471 29404:  120:S ==>     0:  120:R swapper/1
swapper/1           1 1513061754378956639     0:   20:R +   29404:   20: tikv-server
swapper/1           1 1513061754378972019     0:  120:R ==> 29404:  120:R tikv-server
```
