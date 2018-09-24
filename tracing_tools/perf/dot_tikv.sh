#!/bin/bash

set -e

# Usage: sudo dot_tikv.sh pid threads_regex

if [ ! $(which dot) ]; then
  echo "Please install graphviz"
  exit 1
fi

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
  | c++filt\
  | ./gprof2dot.py -f perf -w\
  | dot -Tsvg -o cpu_graph.svg
