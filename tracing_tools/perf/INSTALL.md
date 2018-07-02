## CentOS

```bash
sudo yum -y install unzip
sudo yum -y install perf
wget https://github.com/brendangregg/FlameGraph/archive/master.zip
unzip master.zip
sudo mv FlameGraph-master/ /opt/FlameGraph
```

## Ubuntu

```bash
sudo apt-get -y install unzip
sudo apt-get -y install linux-tools-$(uname -r)
wget https://github.com/brendangregg/FlameGraph/archive/master.zip
unzip master.zip
sudo mv FlameGraph-master/ /opt/FlameGraph
```