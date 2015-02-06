# distelli

> NOTE: THIS STEP IS STILL UNDER HEAVY DEVELOPMENT AND IS NOT READY FOR USE.

Downloads the [Distelli CLI](https://www.distelli.com/docs/distelli-cli-reference) and runs a command.

## Options.

### required

* `command` - command to run. Currently only "supports" `push`.

### optional

* `branches` - a whilelist of branches to allow commands on. If unset, all branches are allowed.
* `manifest` - the distelli manifest file. Required for `push` command.

# Example

Push a build to Distelli. The build will have a commit message of the form 'wercker:${WERCKER_BUILD_ID}`.
This form is necessary to locate the pushed bundle for later deployment.

``` yaml
build:
  steps:
    - distelli:
        command: push
```

# Changelog

## 0.0.1
 - development release
