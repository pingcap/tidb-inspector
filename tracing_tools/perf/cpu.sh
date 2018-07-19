#!/bin/bash

# Usage: sudo cpu.sh pid

perf record -F 99 -p $1 -g -- sleep 60
perf script | /opt/FlameGraph/stackcollapse-perf.pl | /opt/FlameGraph/flamegraph.pl > cpu.svg
