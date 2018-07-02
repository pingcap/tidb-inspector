### Usage
```
$ sudo stap execsnoop-nd.stp
```

Tracing while "ssh pingcap@172.16.20.1" was executed:
```
$ sudo stap execsnoop-nd.stp
TIME                        UID   PPID    PID           COMM ARGS
Mon Dec 11 17:02:08 2017      0  15932  15287           sshd /usr/sbin/sshd -R
Mon Dec 11 17:02:08 2017   1000  15289  15290            zsh -zsh
Mon Dec 11 17:02:08 2017   1000  15290  15291             id /usr/bin/id -un
Mon Dec 11 17:02:08 2017   1000  15290  15292       hostname /usr/bin/hostname
Mon Dec 11 17:02:08 2017   1000  15290  15293             id /usr/bin/id -gn
Mon Dec 11 17:02:08 2017   1000  15290  15294             id /usr/bin/id -un
Mon Dec 11 17:02:08 2017   1000  15290  15295    grepconf.sh /bin/sh /usr/libexec/grepconf.sh -c
Mon Dec 11 17:02:08 2017   1000  15295  15296           grep grep -qsi ^COLOR.*none /etc/GREP_COLORS
Mon Dec 11 17:02:08 2017   1000  15297  15298            tty /usr/bin/tty -s
Mon Dec 11 17:02:08 2017   1000  15290  15297           tput /usr/bin/tput colors
```
