# sysstatsd - A simple graphite system data reporter
A simple application that monitors a local system for CPU, Disk, Memory and Load Average.  It will then output it to a graphite server over UDP.   We built this as a sample platform to interact with Telemetry's agent and showcase its graphite data ingestion capabilities. 

# Installation

This app is designed to work and has been tested on Linux and OSX.  In theory it may run on Windows but this is wholly untested.  

- Install the [OSX Package](https://github.com/telemetryapp/sysstatsd/releases/download/1.0/sysstatsd-osx.tar.gz)
- Install the [Debian/Ubuntu Package](https://github.com/telemetryapp/sysstatsd/releases/download/1.0/sysstatsd_1.0_amd64.deb)

# Building

sysstatsd is written in Go.  In order to build you'll need to have a working Go environment.  Typically we cross compile 

To build for linux:

	GOOS=linux GOARCH=amd64 go build -o pkg/linux_amd64/sysstatsd

# Creating .deb

Usually you'll want to install from a .deb file rather than a binary.  To create a .deb you'll want to do the following:

Install the [FPM Ruby Gem](https://github.com/jordansissel/fpm)

	fpm -s dir -t deb -n 'sysstatsd' -v 1.0 -m support@telemetryapp.com --vendor support@telemetryapp.com --license MIT --url https://telemetryapp.com -e --prefix /bin -C pkg/linux_amd64 sysstatsd
