circonus-webhooks-golang
========================

[![Build Status](https://travis-ci.org/modcloth-labs/circonus-webhooks-golang.svg)](https://travis-ci.org/modcloth-labs/circonus-webhooks-golang)

Generic Circonus Webhook Notification Handlers in Golang (concept borrowed from
https://github.com/circonus-labs/circonus-webhooks-python). Decided to
implement this in Go to receive the benefits of static compilation (reduce
dependencies).

This repository contains custom webhook proxies that allow you to specify this
binary as the webhook URL to forward to other web services in the format that
they understand (e.g. Hipchat).
