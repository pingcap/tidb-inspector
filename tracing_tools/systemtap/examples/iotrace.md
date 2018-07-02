### Usage
```
$ sudo stap iotrace.stp -x <pid>
```

```
$ sudo stap iotrace.stp -x 33709
begin to trace thread IO
prometheus[34730] - open /data3/louis/deploy/data.metrics/bb/062e26428b48dd.db
prometheus[34730] - write 16656 bytes to louis/deploy/data.metrics/bb/062e26428b48dd.db
prometheus[34730] - close /data3/louis/deploy/data.metrics/bb/062e26428b48dd.db
prometheus[33714] - close socket:[368813610]
prometheus[33746] - read 0 bytes from TCP
prometheus[34732] - write 270 bytes to TCP
prometheus[33713] - read 4096 bytes from TCP
prometheus[33713] - read 4096 bytes from TCP
prometheus[33713] - read 4096 bytes from TCP
prometheus[33713] - read 4096 bytes from TCP
prometheus[33713] - read 1530 bytes from TCP
prometheus[33713] - close socket:[368825295]
prometheus[34730] - open /data3/louis/deploy/data.metrics/88/e411a6dcd9e5b7.db
prometheus[34730] - write 2082 bytes to louis/deploy/data.metrics/88/e411a6dcd9e5b7.db
prometheus[34730] - close /data3/louis/deploy/data.metrics/88/e411a6dcd9e5b7.db
```
