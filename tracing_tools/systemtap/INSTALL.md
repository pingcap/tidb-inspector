## CentOS

```
sudo yum install kernel-devel
sudo yum --enablerepo=base-debuginfo install kernel-debuginfo-$(uname -r) kernel-debuginfo-common-$(uname -m)-$(uname -r)
sudo yum -y install systemtap systemtap-runtime
```

## Ubuntu

```
sudo apt-get install -y systemtap gcc
```

- 16.04 and higher
```
sudo apt-key adv --keyserver keyserver.ubuntu.com --recv-keys C8CAB6595FDFF622
```
- older distributions
```
sudo apt-key adv --keyserver keyserver.ubuntu.com --recv-keys ECDCAD72428D7C01
```

```
codename=$(lsb_release -c | awk  '{print $2}')
sudo tee /etc/apt/sources.list.d/ddebs.list << EOF
deb http://ddebs.ubuntu.com/ ${codename}      main restricted universe multiverse
deb http://ddebs.ubuntu.com/ ${codename}-security main restricted universe multiverse
deb http://ddebs.ubuntu.com/ ${codename}-updates  main restricted universe multiverse
deb http://ddebs.ubuntu.com/ ${codename}-proposed main restricted universe multiverse
EOF

sudo apt-get update
sudo apt-get install linux-image-$(uname -r)-dbgsym
```

## Initial Testing
```
sudo stap -v -e 'probe vfs.read {printf("read performed\n"); exit()}'
```