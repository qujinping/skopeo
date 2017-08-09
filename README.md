skopeo [![Build Status](https://travis-ci.org/projectatomic/skopeo.svg?branch=master)](https://travis-ci.org/projectatomic/skopeo)
=

_Please be aware `skopeo` is still work in progress and it currently supports only registry API V2_

`skopeo` is a command line utility for various operations on container images and image repositories.

Inspecting a repository
-
`skopeo` is able to _inspect_ a repository on a Docker registry and fetch images layers.
By _inspect_ I mean it fetches the repository's manifest and it is able to show you a `docker inspect`-like
json output about a whole repository or a tag. This tool, in contrast to `docker inspect`, helps you gather useful information about
a repository or a tag before pulling it (using disk space) - e.g. - which tags are available for the given repository? which labels the image has?

Examples:
```sh
# show properties of fedora:latest
$ skopeo inspect docker://docker.io/fedora
{
    "Name": "docker.io/library/fedora",
    "Tag": "latest",
    "Digest": "sha256:cfd8f071bf8da7a466748f522406f7ae5908d002af1b1a1c0dcf893e183e5b32",
    "RepoTags": [
        "20",
        "21",
        "22",
        "23",
        "heisenbug",
        "latest",
        "rawhide"
    ],
    "Created": "2016-03-04T18:40:02.92155334Z",
    "DockerVersion": "1.9.1",
    "Labels": {},
    "Architecture": "amd64",
    "Os": "linux",
    "Layers": [
        "sha256:236608c7b546e2f4e7223526c74fc71470ba06d46ec82aeb402e704bfdee02a2",
        "sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4"
    ]
}

# show unverifed image's digest
$ skopeo inspect docker://docker.io/fedora:rawhide | jq '.Digest'
"sha256:905b4846938c8aef94f52f3e41a11398ae5b40f5855fb0e40ed9c157e721d7f8"
```

Copying images
-
`skopeo` can copy container images between various storage mechanisms,
e.g. Docker registries (including the Docker Hub), the Atomic Registry,
local directories, and local OCI-layout directories:

```sh
$ skopeo copy docker://busybox:1-glibc atomic:myns/unsigned:streaming
$ skopeo copy docker://busybox:latest dir:existingemptydirectory
$ skopeo copy docker://busybox:latest oci:busybox_ocilayout
```

Deleting images
-
For example,
```sh
$ skopeo delete docker://localhost:5000/imagename:latest
```

Private registries with authentication
-
When interacting with private registries, `skopeo` first looks for `--creds` (for `skopeo inspect|delete`) or `--src-creds|--dest-creds` (for `skopeo copy`) flags. If those aren't provided, it looks for the Docker's cli config file (usually located at `$HOME/.docker/config.json`) to get the credentials needed to authenticate. The ultimate fallback, as Docker does, is to provide an empty authentication when interacting with those registries.

Examples:
```sh
$ cat /home/runcom/.docker/config.json
{
	"auths": {
		"myregistrydomain.com:5000": {
			"auth": "dGVzdHVzZXI6dGVzdHBhc3N3b3Jk",
			"email": "stuf@ex.cm"
		}
	}
}

# we can see I'm already authenticated via docker login so everything will be fine
$ skopeo inspect docker://myregistrydomain.com:5000/busybox
{"Tag":"latest","Digest":"sha256:473bb2189d7b913ed7187a33d11e743fdc2f88931122a44d91a301b64419f092","RepoTags":["latest"],"Comment":"","Created":"2016-01-15T18:06:41.282540103Z","ContainerConfig":{"Hostname":"aded96b43f48","Domainname":"","User":"","AttachStdin":false,"AttachStdout":false,"AttachStderr":false,"Tty":false,"OpenStdin":false,"StdinOnce":false,"Env":null,"Cmd":["/bin/sh","-c","#(nop) CMD [\"sh\"]"],"Image":"9e77fef7a1c9f989988c06620dabc4020c607885b959a2cbd7c2283c91da3e33","Volumes":null,"WorkingDir":"","Entrypoint":null,"OnBuild":null,"Labels":null},"DockerVersion":"1.8.3","Author":"","Config":{"Hostname":"aded96b43f48","Domainname":"","User":"","AttachStdin":false,"AttachStdout":false,"AttachStderr":false,"Tty":false,"OpenStdin":false,"StdinOnce":false,"Env":null,"Cmd":["sh"],"Image":"9e77fef7a1c9f989988c06620dabc4020c607885b959a2cbd7c2283c91da3e33","Volumes":null,"WorkingDir":"","Entrypoint":null,"OnBuild":null,"Labels":null},"Architecture":"amd64","Os":"linux"}

# let's try now to fake a non existent Docker's config file
$ cat /home/runcom/.docker/config.json
{}

$ skopeo inspect docker://myregistrydomain.com:5000/busybox
FATA[0000] unauthorized: authentication required

# passing --creds - we can see that everything goes fine
$ skopeo inspect --creds=testuser:testpassword docker://myregistrydomain.com:5000/busybox
{"Tag":"latest","Digest":"sha256:473bb2189d7b913ed7187a33d11e743fdc2f88931122a44d91a301b64419f092","RepoTags":["latest"],"Comment":"","Created":"2016-01-15T18:06:41.282540103Z","ContainerConfig":{"Hostname":"aded96b43f48","Domainname":"","User":"","AttachStdin":false,"AttachStdout":false,"AttachStderr":false,"Tty":false,"OpenStdin":false,"StdinOnce":false,"Env":null,"Cmd":["/bin/sh","-c","#(nop) CMD [\"sh\"]"],"Image":"9e77fef7a1c9f989988c06620dabc4020c607885b959a2cbd7c2283c91da3e33","Volumes":null,"WorkingDir":"","Entrypoint":null,"OnBuild":null,"Labels":null},"DockerVersion":"1.8.3","Author":"","Config":{"Hostname":"aded96b43f48","Domainname":"","User":"","AttachStdin":false,"AttachStdout":false,"AttachStderr":false,"Tty":false,"OpenStdin":false,"StdinOnce":false,"Env":null,"Cmd":["sh"],"Image":"9e77fef7a1c9f989988c06620dabc4020c607885b959a2cbd7c2283c91da3e33","Volumes":null,"WorkingDir":"","Entrypoint":null,"OnBuild":null,"Labels":null},"Architecture":"amd64","Os":"linux"}

# skopeo copy example:
$ skopeo copy --src-creds=testuser:testpassword docker://myregistrydomain.com:5000/private oci:local_oci_image
```
If your cli config is found but it doesn't contain the necessary credentials for the queried registry
you'll get an error. You can fix this by either logging in (via `docker login`) or providing `--creds` or `--src-creds|--dest-creds`.

Building
-
To build the `skopeo` binary you need at least Go 1.5 because it uses the latest `GO15VENDOREXPERIMENT` flag.

There are two ways to build skopeo: in a container, or locally without a container.  Choose the one which better matches your needs and environment.

### Building without a container
Building without a container requires a bit more manual work and setup in your environment, but it is more flexible:
- It should work in more environments (e.g. for native macOS builds)
- It does not require root privileges (after dependencies are installed)
- It is faster, therefore more convenient for developing `skopeo`.

Install the necessary dependencies:
```sh
Fedora$ sudo dnf install gpgme-devel libassuan-devel btrfs-progs-devel device-mapper-devel
macOS$ brew install gpgme
```

Make sure to clone this repository in your `GOPATH` - otherwise compilation fails.

```sh
$ git clone https://github.com/projectatomic/skopeo $GOPATH/src/github.com/projectatomic/skopeo
$ cd $GOPATH/src/github.com/projectatomic/skopeo && make binary-local
```

### Building in a container
Building in a container is simpler, but more restrictive:
- It requires the `docker` command and the ability to run Linux containers
- The created executable is a Linux executable, and depends on dynamic libraries which may only be available only in a container of a similar Linux distribution.

```sh
$ make binary # Or (make all) to also build documentation, see below.
```

### Building documentation
To build the manual you will need go-md2man.
```sh
Debian$ sudo apt-get install go-md2man
Fedora$ sudo dnf install go-md2man
```
Then
```sh
$ make docs
```

Installing
-
If you built from source:
```sh
$ sudo make install
```
`skopeo` is also available from Fedora 23 (and later):
```sh
$ sudo dnf install skopeo
```
TODO
-
- list all images on registry?
- registry v2 search?
- show repo tags via flag or when reference isn't tagged or digested
- support rkt/appc image spec

NOT TODO
-
- provide a _format_ flag - just use the awesome [jq](https://stedolan.github.io/jq/)

CONTRIBUTING
-

### Dependencies management

`skopeo` uses [`vndr`](https://github.com/LK4D4/vndr) for dependencies management.

In order to add a new dependency to this project:

- add a new line to `vendor.conf` according to `vndr` rules (e.g. `github.com/pkg/errors master`)
- run `vndr github.com/pkg/errors`

In order to update an existing dependency:

- update the relevant dependency line in `vendor.conf`
- run `vndr github.com/pkg/errors`

When new PRs for [containers/image](https://github.com/containers/image) break `skopeo` (i.e. `containers/image` tests fail in `make test-skopeo`):

- create out a new branch in your `skopeo` checkout and switch to it
- update `vendor.conf`. Find out the `containers/image` dependency; update it to vendor from your own branch and your own repository fork (e.g. `github.com/containers/image my-branch https://github.com/runcom/image`)
- run `vndr github.com/containers/image`
- make any other necessary changes in the skopeo repo (e.g. add other dependencies now requied by `containers/image`, or update skopeo for changed `containers/image` API)
- optionally add new integration tests to the skopeo repo
- submit the resulting branch as a skopeo PR, marked “DO NOT MERGE”
- iterate until tests pass and the PR is reviewed
- then the original `containers/image` PR can be merged, disregarding its `make test-skopeo` failure
- as soon as possible after that, in the skopeo PR, restore the `containers/image` line in `vendor.conf` to use `containers/image:master`
- run `vndr github.com/containers/image`
- update the skopeo PR with the result, drop the “DO NOT MERGE” marking
- after tests complete succcesfully again, merge the skopeo PR

License
-
skopeo is licensed under the Apache License, Version 2.0. See
[LICENSE](LICENSE) for the full license text.
