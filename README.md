# .

[![GoDoc](https://img.shields.io/badge/pkg.go.dev-doc-blue)](http://pkg.go.dev/github.com/davexre/sitescan)

Package sitescan is a basic web scraping tool that compares two file trees,
and prints out the differences.

sitescan can be configured in several different ways. At a minimum, it needs
to be told the correct URLs to visit. It can also handle basic HTTP authentication
(username and password). Optionally, you can specify a friendlier name for
each site, as well. Because it uses Viper for configuration processing, sitescan
is very flexible in terms of how to configure it. It will accept a YAML based
command line options, environment variables, and config files - or a combination of
all three. Precedence is as listed.

Command Line Usage:

```diff
-c, --config string      path to alternate configuration file
-d, --debug              output debugging info
    --site1 string       Site 1 URL
    --site1name string   Site 1 Name
    --site1pass string   Site 1 Password
    --site1user string   Site 1 User ID
    --site2 string       Site 2 URL
    --site2name string   Site 2 Name
    --site2pass string   Site 2 Password
    --site2user string   Site 2 User ID
```

## Environment Variables

Acceptable environment variables are all capitals, are prefixed with "SITESCAN_",
and otherwise match the command line switches:

```go
SITESCAN_SITE1
SITESCAN_SITE1NAME
SITESCAN_SITE1PASS
SITESCAN_SITE1USER
SITESCAN_SITE2
SITESCAN_SITE2NAME
SITESCAN_SITE2PASS
SITESCAN_SITE2USER
```

## Config File

The default configuration file is named "sitescan_config.yaml" and should reside
in the directory you're running sitescan from (i.e. the directory that sitescan
will see as "PWD"). You can specify an alternate config file name/path using the
-c / --config command line option. And example config file:
`	# Example sitescan_config.yaml file

```go
site1: [http://webserver.myhost.com/path/to/examine](http://webserver.myhost.com/path/to/examine)
site2: [http://www.anotherhost.org:8080/](http://www.anotherhost.org:8080/)
site1user: someguy
site1pass: spaceballs12345
site1name: MyHost.com site
# site2user:
# site2pass:
site2name: AnotherHost site `
```

---
Readme created from Go doc with [goreadme](https://github.com/posener/goreadme)
