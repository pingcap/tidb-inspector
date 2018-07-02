#!/bin/bash

# Usage: sudo io.sh pid
#
# Note: 
# You may need to rebuild your kernel to support kernel scheduler statistics and then 
# enable it with `sudo echo 1 | sudo tee /proc/sys/kernel/sched_schedstats`

perf record -e sched:sched_stat_sleep -e sched:sched_switch \
    -e sched:sched_process_exit -p $1 -g -o perf.data.raw sleep 10
perf inject -v -s -i perf.data.raw -o perf.data
perf script -F comm,pid,tid,cpu,time,period,event,ip,sym,dso,trace | awk '
    NF > 4 { exec = $1; period_ms = int($5 / 1000000) }
    NF > 1 && NF <= 4 && period_ms > 0 { print $2 }
    NF < 2 && period_ms > 0 { printf "%s\n%d\n\n", exec, period_ms }' | \
    /opt/FlameGraph/stackcollapse.pl | \
    /opt/FlameGraph/flamegraph.pl --countname=ms --title="Off-CPU Time Flame Graph" --colors=io > offcpu.svg
