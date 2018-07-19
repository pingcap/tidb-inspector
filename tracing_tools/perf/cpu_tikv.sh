#!/bin/bash

# Usage: sudo cpu_tikv.sh pid

if [ "$#" -ne 1 ]; then
  PID=`pgrep tikv-server | head -n 1`
  echo "Note: PID is not specified. Capturing for tikv-server PID = $PID"
else
  PID=$1
fi

perf record -F 99 -p $PID --call-graph=dwarf sleep 60
perf script | ./fold-tikv-threads-perf.pl | /opt/FlameGraph/stackcollapse-perf.pl | /opt/FlameGraph/flamegraph.pl > cpu.svg
