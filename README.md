# sysstatsd
A simple application that monitors a local system for CPU, Disk, Memory and Load Average.  It will then output it to a graphite server.
# Linux Build

To build for linux:

GOOS=linux GOARCH=amd64 go build -o pkg/linux_amd64/sysstatsd

# Using Goconvey

go get -t github.com/smartystreets/goconvey
goconvey
visit http://localhost:8080

# Creating .deb

gem install fpm # First time only https://github.com/jordansissel/fpm
fpm -s dir -t deb -n 'sysstatsd' -v 1.0 -m support@telemetryapp.com --vendor support@telemetryapp.com --license MIT --url https://telemetryapp.com -e --prefix /bin -C pkg/linux_amd64 sysstatsd
