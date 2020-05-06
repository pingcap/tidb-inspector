#!/bin/bash

throttle_cfg_file=/proc/sys/kernel/perf_cpu_time_max_percent
original_throttle_cfg=`cat $throttle_cfg_file`

# Disable the mechanism. Do not monitor or correct perf's sampling rate 
# no matter how CPU time it takes.
echo 0 > $throttle_cfg_file

output=`date '+%T'`-c2c.txt
# Run “perf c2c” for 3, 5, or 10 seconds. Running it any longer may take
# you from seeing concurrent false sharing to seeing cacheline accesses
# which are disjoint in time.
perf c2c record -F 60000 -a --all-user --call-graph dwarf sleep 5

perf c2c report -NN -c tid,iaddr --full-symbols --source --stdio > "$output"

# Recover sampling rate limit
echo $original_throttle_cfg > /proc/sys/kernel/perf_cpu_time_max_percent
