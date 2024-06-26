# 2.25.0 / 2024-05-04

  * Add float64 field type support

# 2.24.0 / 2024-04-03

  * Add validation to fieldcollection

# 2.23.0 / 2024-03-06

  * Add `ThrottledReader` as io-helper
  * Update dependencies

# 2.22.0 / 2023-11-27

  * Add special error to terminate retries immediately

# 2.21.0 / 2023-10-23

  * Replace and deprecate `float.Round` method

# 2.20.1 / 2023-10-12

  * Remove gomega from tests and update dependencies

# 2.20.0 / 2023-06-17

  * Add CSP helper

# 2.19.0 / 2023-06-16

  * Add `http.NoListFS`
  * Update dependencies, fix multiple CVEs

# 2.18.0 / 2023-06-10

  * Add file.FSStack implementation

# 2.17.1 / 2023-05-19

  * Fix: Prevent panics when no arguments are given

# 2.17.0 / 2023-05-19

  * Add simple CLI helper

# 2.16.0 / 2023-03-19

  * Allow to set watcher to follow symlinks
  * Drop support for Go 1.18 in tests

# 2.15.3 / 2023-03-18

  * Fix: Tests broken after last change

# 2.15.2 / 2023-03-18

  * Fix logic bug in run loop, replace Stat with Lstat

# 2.15.1 / 2023-03-07

  * Update dependencies

# 2.15.0 / 2023-02-06

  * Add `http.LogRoundTripper` helper for request debugging

# 2.14.0 / 2023-01-28

  * Add `file.Watcher` helper
  * [ci] Add test as Github workflow

# 2.13.0 / 2021-11-20

  * Add `fieldcollection` helper

# 2.12.2 / 2021-03-09

  * Fix: Do not panic on weird env list entries

# 2.12.1 / 2021-02-06

  * Fix: Pass in logger

# 2.12.0 / 2021-02-06

  * Update dependencies
  * Allow to pass in a logger for HTTP logs
  * Update imports to v2 import paths

# 2.11.0 / 2020-08-07

  * Add convenience wrapper around property sets
  * Drop support for Go <1.13
  * Add test for successful execution

# 2.10.0 / 2019-11-15

  * Add backoff retry-helper

# 2.9.1 / 2019-02-28

  * Fix unversioned import paths

# 2.9.0 / 2019-02-28

  * Add support for Go 1.11+ modules

# 2.8.1 / 2018-11-19

  * Also log query parameters

# 2.8.0 / 2018-09-17

  * Add GZip wrapper

# 2.7.0 / 2018-07-05

  * Add helpers to parse time strings using multiple formats at once

# 2.6.0 / 2018-06-07

  * Add a YAML to JSON converter as yaml-helper

# 2.5.0 / 2018-04-23

  * Add output splitter

# 2.4.0 / 2018-04-03

  * Add proxy IP detection

# 2.3.1 / 2017-11-05

  * Fix TIP version error: Sprintf format %s has arg of wrong type byte
  * Travis: Test on Go 1.7, 1.8, 1.9, tip

# 2.3.0 / 2017-11-05

  * Implement digest header generation

# 2.2.0 / 2017-04-13

  * Add HTTPLogHandler

# 2.1.0 / 2016-12-23

  * Add time.Duration formatter

# 2.0.0 / 2016-10-12

  * Drop Go1.5 / Go1.6 support with using contexts
  * Add github-binary update helper

# 1.4.0 / 2016-05-29

  * Added environment helpers

# 1.3.0 / 2016-05-18

  * Added AccessLogResponseWriter

# 1.2.0 / 2016-05-16

  * Added helper to find binaries in path or directory

# 1.1.0 / 2016-05-06

  * Added Haversine helper functions


1.0.0 / 2016-04-23
==================

  * First versioned revision for use with gopkg.in
