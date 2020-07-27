# gstransfer [WIP]

Google Cloud Storage upload relay/gateway with destination path rewrite support & zero disk allocation, **unstable** version.

## Installation

Currently we only provide .deb package, [get it here](https://github.com/codenoid/gstransfer/releases), or [Build from source](#building-from-source)

## Usage

there is two part of gstransfer, server and client/uploader

```bash
define -bind flag if you want to use server mode
Usage of ./gstransfer:
  -bind string
    	set gstransfer bind ip and port
  -bucket string
    	your bucket name
  -db-path string
    	full path to database folder, will automatically craete if not exist (default "/tmp/gstransfer")
  -dst-rule string
    	dir-0 is root of chosen directory, and so on, and 'filename' is your file original filename, you can add custom prefix (default "dir-0/.../filename")
  -id string
    	define process id, this will enable re-upload without duplicate file on app crash
  -max-size int
    	size in MB (default 100)
  -server string
    	your server url
  -source string
    	full path to your source directory
```

### Server

For server, make sure you had access to bucket target, define GOOGLE_APPLICATION_CREDENTIALS env var with path to .json auth file, for more, read [Getting started with authentication](https://cloud.google.com/docs/authentication/getting-started)

```bash
$ # start gstransfer server
$ gstransfer -bind 0.0.0.0:3000 -max-size 150
```

### Client/Uploader

```bash
$ ./gstransfer -source /full/path/to/dir -bucket name -server http://gstorage-server:8003 -id test-1 -dst-rule dir-0/../file
```

**-dst-rule explanation**

for example you have this full path, `/home/codenoid/Documents/codenoid/gstransfer/` with file located in `/home/codenoid/Documents/codenoid/gstransfer/.git/hooks/post-update.sample`, let's note the **`.git/hooks/post-update.sample`**

1. `-dst-rule dir-0/file` become `.git/post-update.sample` object
2. `-dst-rule dir-0/../file` become `.git/hooks/post-update.sample` object
3. `-dst-rule dir-1/../file` become `hooks/post-update.sample` object
4. `-dst-rule dir-4/../file` become error, because current listed file not match with the rule
5. `-dst-rule file` become `post-update.sample` object

## Building from source

You need to install Go, for linux, [read here](https://codenoid.github.io/posts/cara-install-golang-di-linux/)

```bash
git clone https://github.com/codenoid/gstransfer.git
cd gstransfer
go build -trimpath
```

## FAQ

Q: What if the app crash/exit in the middle progress ?
A: don't worry, just define your existing -id, for recovering from server-crash put `-db-path` anywhere besides `/tmp`

Q: What good server spec for the server mode ?
A: most of my file are lower than 50MB, i can just go with GCP `f1-micro`
