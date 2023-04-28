# gping!

This is an unimaginative golang command-line terminal user interface
application that allows you to continuously ping multiple hosts.
It's really neat!

# Quick Docker Usage

If you want to spin up a container quickly on docker just use the 
`pwhelan/gping:latest` container:

```bash
docker run --rm -ti pwhelan/gping:latest
```

# Installation

To install the application, simply clone the repository to your local machine
and then build it:

```bash
git clone https://github.com/pwhelan/gping
cd gping
go build ./
```

# Usage

This could not be any simpler, just simply invoke gping with a list of IP
addresses and hostnames to ping:

![gping demo](demo.gif)

# Contributing

Contributions to this project are welcome. To contribute, simply fork the
repository, make your changes, and submit a pull request.

# License

This project is licensed under the MIT license - see the LICENSE.md file for
details.
