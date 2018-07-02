SystemTap Tools
===========================

## Installing

See [INSTALL.md](INSTALL.md) for installation steps on your platform.

## Tools

### Scheduler

- [cstop_pid.stp](tools/cstop_pid.stp): Trace top 20 context switch pid. [Examples](examples/cstop_pid.md)
- [cstop_tid.stp](tools/cstop_tid.stp): Trace top 20 context switch tid. [Examples](examples/cstop_tid.md)
- [sched_switch.stp](tools/sched_switch.stp): Trace the scheduler switches happening with the process. [Examples](examples/sched_switch.md)

### IO

- [iotrace.stp](tools/iotrace.stp): Trace thread IO. [Examples](examples/iotrace.md)

### Syscall

- [execsnoop-nd.stp](tools/execsnoop-nd.stp): Trace process `exec()` with command line argument details. [Examples](examples/execsnoop-nd.md)
- [killsnoop-nd.stp](tools/killsnoop-nd.stp): Trace `kill()` signals showing process and signal details. [Examples](examples/killsnoop-nd.md)
