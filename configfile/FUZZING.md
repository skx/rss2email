# Fuzz-Testing

The upcoming 1.18 release of the golang compiler/toolset has integrated
support for fuzz-testing.

Fuzz-testing is basically magical and involves generating new inputs "randomly"
and running test-cases with those inputs.


## Running

If you're running 1.18beta1 or higher you can run the fuzz-testing against
our configuration-file parser like so:

    $ go test -fuzztime=300s -parallel=1 -fuzz=FuzzParser -v
    ..
    --- PASS: TestFuzz (0.00s)
    === FUZZ  FuzzParser
    fuzz: elapsed: 0s, gathering baseline coverage: 0/131 completed
    fuzz: elapsed: 0s, gathering baseline coverage: 131/131 completed, now fuzzing with 1 workers
    fuzz: elapsed: 3s, execs: 21154 (7050/sec), new interesting: 0 (total: 126)
    fuzz: elapsed: 6s, execs: 42006 (6951/sec), new interesting: 0 (total: 126)
    fuzz: elapsed: 9s, execs: 62001 (6663/sec), new interesting: 0 (total: 126)
    ...
    ...
    fuzz: elapsed: 4m57s, execs: 1143698 (0/sec), new interesting: 7 (total: 156)
    fuzz: elapsed: 5m0s, execs: 1143698 (0/sec), new interesting: 7 (total: 156)
    fuzz: elapsed: 5m1s, execs: 1143698 (0/sec), new interesting: 7 (total: 156)
    --- PASS: FuzzParser (301.07s)
    PASS
    ok  	github.com/skx/rss2email/configfile	301.135s


You'll note that I've added `-parellel=1` to the test, because otherwise my desktop system becomes unresponsive while the testing is going on.
