#!/bin/bash

# Usage: sudo cpu_tikv.sh pid threads_regex

if [ -z "$*" ]; then
  PID=`pgrep tikv-server | head -n 1`
  echo "Note: PID is not specified. Capturing for tikv-server PID = $PID"
else
  PID=$1
fi

[ -z $2 ] && THREADS=".*" || THREADS="$2"

perf record -F 99 -p $PID --call-graph=dwarf sleep 60
perf script\
  | ./fold-tikv-threads-perf.pl --threads "$THREADS"\
  | /opt/FlameGraph/stackcollapse-perf.pl\
  | /opt/FlameGraph/flamegraph.pl > cpu.svg
