

[![license](http://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/udhos/jazigo/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/udhos/jazigo)](https://goreportcard.com/report/github.com/udhos/jazigo)
[![Go Reference](https://pkg.go.dev/badge/github.com/udhos/jazigo.svg)](https://pkg.go.dev/github.com/udhos/jazigo)

Table of Contents
=================

* [About Jazigo](#about-jazigo)
* [Supported Platforms](#supported-platforms)
* [Features](#features)
* [Requirements](#requirements)
* [Building and Installing](#building-and-installing)
* [Quick Start - Short version](#quick-start---short-version)
* [Quick Start - Detailed version](#quick-start---detailed-version)
* [Global Settings](#global-settings)
* [Importing Many Devices](#importing-many-devices)
* [SSH Ciphers](#ssh-ciphers)
* [Using AWS S3](#using-aws-s3)
* [Calling an external program](#calling-an-external-program)

Created by [gh-md-toc](https://github.com/ekalinin/github-markdown-toc.go)

About Jazigo
============

Jazigo is a network device configuration backup tool. It supports SSH, TELNET and HTTP(S) protocols.

Supported Platforms
===================

* Linux
* Windows (experimental)
* macOS (experimental)

Features
========

- Backup Cisco, Juniper, Arista, HP Procurve, Mikrotik, Ubiquiti EdgeSwitch, Dell Force10, Brocade VDX and other devices.
- Store backup files in a local directory or AWS S3 bucket.
- View file differences directly from the web UI.
- Support for SSH and TELNET.
- Can directly store backup files into AWS S3 bucket.
- Can call an external program and collect its output.

Requirements
============

- You need a [system with the Go language](https://golang.org/dl/) version 1.22 or later in order to build the application. There is no special requirement for running it.

Building and Installing
======================

To build and install Jazigo, follow these steps:

```bash
# Clone the repository (outside of GOPATH)
git clone https://github.com/udhos/jazigo

# Change directory
cd jazigo

# Run the build script to check for issues and install dependencies
./build.sh

# Install the application
go install ./jazigo
```

Quick Start - Short version
===========================

This is how to boot up Jazigo very quickly:

```bash
# Clone the repository (outside of GOPATH)
git clone https://github.com/udhos/jazigo

# Change directory and create necessary directories
cd jazigo
mkdir etc repo log

# Install and run the application
JAZIGO_HOME=$PWD go install ./jazigo
$GOPATH/bin/jazigo
```

Open Jazigo interface - http://localhost:8080/jazigo/

Quick Start - Detailed version
==============================

This is how to boot up Jazigo very quickly:

```bash
# Clone the repository (outside of GOPATH)
git clone https://github.com/udhos/jazigo

# Change directory and create necessary directories
cd jazigo
mkdir etc repo log

# Install and run the application
JAZIGO_HOME=$PWD go install ./jazigo
$GOPATH/bin/jazigo
```

Open Jazigo interface - http://localhost:8080/jazigo/

Global Settings
===============

You might want to adjust global settings. See the Jazigo *admin* window under [http://localhost:8080/jazigo/admin](http://localhost:8080/jazigo/admin).

    maxconfigfiles: 120
    holdtime: 12h0m0s
    scaninterval: 10m0s
    maxconcurrency: 20
    maxconfigloadsize: 10000000

**maxconfigfiles**: This option limits the amount of files stored per device. When this limit is reached, older files are discarded.

**holdtime**: When a successful backup is saved for a device, the software will only contact that specific device again *after* expiration of the 'holdtime' timer.

**scaninterval**: The interval between two device table scans. If the device table is fully processed before the 'scaninterval' timer, the software will wait idly for the next scan cycle. If the full table scan takes longer than 'scaninterval', the next cycle will start immediately.

**maxconcurrency**: This option limits the number of concurrent backup jobs. You should raise this value if you need faster scanning of all devices. Keep in mind that if your devices use a centralized authentication system (for example, Cisco Secure ACS), the authentication server might become a bottleneck for high concurrency.

**maxconfigloadsize**: This limit puts restriction into the amount of data the tool loads from a file to memory. Intent is to protect the servers' memory from exhaustion while trying to handle multiple very large configuration files.

Importing Many Devices
======================

You can import many devices at once by using the following command:

```bash
$GOPATH/bin/jazigo -import /path/to/device-list.csv
```

The CSV file should have the following format:

```
hostname,protocol,username,password,port,enable_password
router1,ssh,user,pass,,enpass
switch1,telnet,user,pass,23,
```

SSH Ciphers
===========

You can specify which ciphers to use for SSH connections by setting the `SSH_CIPHERS` environment variable. For example:

```bash
export SSH_CIPHERS=aes128-ctr,aes192-ctr,aes256-ctr
$GOPATH/bin/jazigo
```

Using AWS S3
===========

To store backup files in an AWS S3 bucket, you need to configure the following settings:

```bash
export JAZIGO_S3_REGION=us-west-2
export JAZIGO_S3_BUCKET=my-bucket
$GOPATH/bin/jazigo
```

Calling an external program
===========================

You can call an external program and collect its output by using the `exec` protocol. For example:

```
hostname,protocol,username,password,port,enable_password
router1,exec,user,pass,,enpass,/path/to/program,program arguments
```

