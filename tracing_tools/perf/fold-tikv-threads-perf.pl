#!/usr/bin/perl -w
#
# Fold TiKV thread pools threads into one for perf output.
#
# Usage: ./fold-tikv-threads-perf.pl infile > outfile

use strict;
use Getopt::Long;

my $threads = ""; # regex of thread names.
my $help = 0;

sub usage {
  die <<USAGE_END;
USAGE: $0 [options] infile > outfile.svg\n
  --threads REGEX  # captrue thread names.
  --help           # this message

  eg,
  $0 --threads="grpc.*" infile > outfile
USAGE_END
}

GetOptions(
  'threads=s'  => \$threads,
  'help'        => \$help,
) ? ($help && usage()) : usage();

my $threads_regex = qr/$threads/;
my $skip_thread = 0;

while (my $line = <>) {
  chomp $line;
  if ($line =~ /^(\S.+?)\s+(\d+)(.*)$/) {
    my $command = $1;
    my $pid = $2;
    my $remain = $3;

    if ($threads) {
      # We want to filter some threads.
      if ($command =~ /$threads_regex/) {
        $skip_thread = 0;
      } else {
        $skip_thread = 1;
        # Skips this line too.
        next;
      }
    }

    # grpc server. e.g. grpc-server-0
    $command =~ s/^grpc-server-\d*$/grpc-server/;

    # coprocessor read pool. e.g. cop-high0, cop-normal0, cop-low0
    $command =~ s/^cop-high\d*$/cop-high/;
    $command =~ s/^cop-normal\d*$/cop-normal/;
    $command =~ s/^cop-low\d*$/cop-low/;

    # raftstore. e.g. raftstore-1
    $command =~ s/^raftstore-\d*$/raftstore/;
    # raftstore is renamed to raftstore-1-0 in 3.0
    $command =~ s/^raftstore-\d*-\d*$/raftstore/;

    # SST importer. e.g. sst-importer0
    $command =~ s/^sst-importer\d*$/sst-importer/;

    # storage read pool. e.g. store-read-high, store-read-norm, store-read-low0
    # ony low\d needs to be unified, because other priorities already have same thread names because of truncation.
    $command =~ s/^store-read-low\d*$/store-read-low/;

    # rocksdb. e.g. rocksdb:bg0 rocksdb:low0 rocksdb:high0
    $command =~ s/^rocksdb:bg\d*$/rocksdb:bg/;
    $command =~ s/^rocksdb:low\d*$/rocksdb:low/;
    $command =~ s/^rocksdb:high\d*$/rocksdb:high/;

    # snapshot sender. e.g. snap sender0
    $command =~ s/^snap sender\d*$/snap-sender/;
    # thread is renamed to snap-sender0 in newer versions
    $command =~ s/^snap-sender\d*$/snap-sender/;

    # apply. e.g. apply-1
    $command =~ s/^apply-\d*$/apply/;

    # future-poller. e.g. futurue-poller-0
    $command =~ s/^future-poller-\d*$/future-poller/;

    # backup. e.g. backup-worker0
    $command =~ s/^backup-worker\d*$/backup-worker/;

    # sst importer. e.g. sst-importer0
    $command =~ s/^sst-importer\d*$/sst-importer/;

    # CDC. e.g. cdcwkr0
    $command =~ s/^cdcwkr\d*$/cdcwkr/;

    $line = $command . " " . $pid . $remain;
  } elsif ($skip_thread) {
    # The thread is skipped.
    next;
  }

  print $line, "\n";
}
