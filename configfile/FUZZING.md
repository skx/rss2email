# Fuzz-Testing

If you don't have the appropriate tools installed you can fetch them via:

    $ go get github.com/dvyukov/go-fuzz/go-fuzz
    $ go get github.com/dvyukov/go-fuzz/go-fuzz-build

Now you can build this package with fuzzing enabled:

    $ go-fuzz-build .

Create a location to hold the work, and give it at least one input:

    $ mkdir -p workdir/corpus
    $ cat > workdir/corpus/1.txt<<EOF
    https://example.com/
     - option:value
    # Comment
    https://example.org/
    # Comment
    EOF
    $ echo "https://example.net/" > workdir/corpus/2.txt
    $ echo "# https://example.net/" > workdir/corpus/3.txt

Now you can actually launch the fuzzer - here I use `-procs 1` so that
my desktop system isn't complete overloaded:

    $ go-fuzz -procs 1 -bin=configfile-fuzz.zip -workdir workdir/
