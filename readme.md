lncount
--------

Simple and fast tool to count how many lines your files have.

Features
-------

-	Supports glob patterns.
-	Reads from stdin when piped.
-	Concurrent and fast.

Some Examples
-------

	lncount **/*.go #read all .go files under the current directory
	cat main.go | lncount #reads from stdin
	lncount main.go # reads just main.go
	lncount **/main.go # only count main.go files

Installation
--------

You need the go compiler.

	go get -u -v github.com/insomnimus/lncount

This should install lncount to your $GOBIN directory, make sure $GOBIN is in your path.