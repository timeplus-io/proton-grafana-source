# Changelog

## 2.0.0

Key changes:

* Removed "AddNow" option
* Fixed boolean field issue
* Refactored code based on Grafana Go SDK v0.251
* Updated proton go driver to 2.0.17

## 1.0.3

Signed and approved by Grafana Inc.

Key changes:

* No longer need to specify the query is streaming or not. Call Proton query analazyer (need port 3218 open)
* Support for dashboard variables
* Filter query if the SQL is empty or disabled
* Enable Grafana Alerting

Please note you need to enable both 8463 and 3218 ports from Timeplus Proton.

## 1.0.0 (Unreleased)

Initial release.
