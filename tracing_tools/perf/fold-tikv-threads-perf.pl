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
    $command =~ s/^grpc-server-.*$/grpc-server/;
    $command =~ s/^cop-.*$/cop/;
    $command =~ s/^raftstore-.*$/raftstore/;
    $command =~ s/^store-read-.*$/store-read/;
    $command =~ s/^rocksdb:bg.*$/rocksdb:bg/;
    print $command, " ", $pid, $remain, "\n";
  } else {
    # other
    print $_, "\n";
  }
}
