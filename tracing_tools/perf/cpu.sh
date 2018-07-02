#!/bin/bash

# Usage: sudo cpu.sh pid

perf record -F 99 -p $1 -g -- sleep 60
perf script > out.perf
/opt/FlameGraph/stackcollapse-perf.pl out.perf > out.folded
/opt/FlameGraph/flamegraph.pl out.folded > cpu.svg
