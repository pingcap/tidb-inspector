#!/usr/bin/perl -w
#
# Fold TiKV thread pools threads into one for perf output.
#
# Usage: ./fold-tikv-threads-perf.pl infile > outfile

while (defined($_ = <>)) {
  chomp;
  if (/^(\S.+?)\s+(\d+)(.*)$/) {
    my $command = $1;
    my $pid = $2;
    my $remain = $3;
    # grpc server. e.g. grpc-server-0
    $command =~ s/^grpc-server-\d*$/grpc-server/;

    # coprocessor read pool. e.g. cop-high0, cop-normal0, cop-low0
    $command =~ s/^cop-high\d*$/cop-high/;
    $command =~ s/^cop-normal\d*$/cop-normal/;
    $command =~ s/^cop-low\d*$/cop-low/;

    # raftstore. e.g. raftstore-1
    $command =~ s/^raftstore-\d*$/raftstore/;

    # SST importer. e.g. sst-importer0
    $command =~ s/^sst-importer\d*$/sst-importer/;

    # storage read pool. e.g. store-read-high, store-read-norm, store-read-low0
    # ony low\d needs to be unified, because other priorities already have same thread names because of truncation.
    $command =~ s/^store-read-low\d*$/store-read-low/;

    # rocksdb. e.g. rocksdb:bg0
    $command =~ s/^rocksdb:bg\d*$/rocksdb:bg/;

    # snapshot sender. e.g. snap sender0
    $command =~ s/^snap sender\d*$/snap sender/;

    print $command, " ", $pid, $remain, "\n";
  } else {
    # other
    print $_, "\n";
  }
}
