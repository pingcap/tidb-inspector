#!/bin/bash

# Usage: sudo mem.sh pid
# 
# Note:
#   This script can collect the malloc profile, if you link other memory allocator
#   like jemalloc, you may use `perf probe` to add the probe at first.
#   E.g., for TiKV, we use `sudo perf probe -x tikv_binary -a malloc`, and we also 
#   need to use `perf record -e probe_tikv:malloc -F 99 -p $1 -g -- sleep 10` instead.

perf record -e malloc -F 99 -p $1 -g -- sleep 10
perf script > out.perf
/opt/FlameGraph/stackcollapse-perf.pl out.perf > out.folded
/opt/FlameGraph/flamegraph.pl  --colors=mem out.folded > mem.svg