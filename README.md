# Stash - simple http file storage

# How it works

Stashd is a simple http server that presents a secured hash-indexed file storage.
The client is a thin wrapper around the curl program to provide convenient
access from the command line.
File names are not stored by design.

Example client usage:
```bash
# First we create a file
$ echo "Hello stashd! I am a file!" > testfile

# Now we put the file into the stash, which returns our file id
$ ./stash put testfile
67a54debf93e99a5887d1e60c7dadf2ce2a2a970

# Using this id, we can get the contents of the file stored online
$ ./stash get 67a54debf93e99a5887d1e60c7dadf2ce2a2a970
Hello stashd! I am a file!

# Now we can delete the file
$ ./stash delete 67a54debf93e99a5887d1e60c7dadf2ce2a2a970
ok

# After deleting a file, it's contents are lost
$ ./stash get 67a54debf93e99a5887d1e60c7dadf2ce2a2a970
error reading or finding file
```

# Building

Stashd requires gorilla mux. So to build stashd and its dependencies
simply, run:

```bash
$ go get github.com/gorilla/mux
$ cd stashd
$ go build
```


# Starting the server

First, create a random secret, e.g. using OpenSSL, and store it somewhere:
```bash
$ openssl rand -hex 12 > ~/.stash-secret
```

Next, start the server

```bash
$ mkdir /tmp/stashd-data
$ stashd --dir /tmp/stashd-data --secret ~/.stash-secret \
  --cert <path-to-ssl-certificate> --key <path-to-ssl-key>
```

Done. If you want to run the server using a prefix e.g.
`yourdomain.tld/stash/`, start the server with the option `--prefix /stash`.

# Configuring the client

The client needs two pieces of configuration:

1. The server secret (required for putting and deleting files). The server
   secret needs to be stored in `~/.stash-secret`.

2. The server url, which needs to be configured in the script directly in line 4.
