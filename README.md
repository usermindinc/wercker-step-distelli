# distelli

Downloads the [Distelli CLI](https://www.distelli.com/docs/distelli-cli-reference) and runs a command.

## Options.

### required

* `command` - command to run. Currently only "supports" `push` and `deploy`.
* `accessKey` - the Distelli access key (token) for using the CLI
* `secretKey` - the Distelli secret key for using the CLI

### optional

* `branches` - a whilelist of branches to allow commands on. If unset, all branches are allowed.
* `manifest` - the distelli manifest file. Required for `push` command.

# Example

Push a build to Distelli. The build will have a commit message of the form `wercker:${WERCKER_BUILD_ID}`.
This form is necessary to locate the pushed bundle for later deployment.

``` yaml
build:
  steps:
    - distelli:
        accessKey: ${DISTELLI_TOKEN}
        secretKey: ${DISTELLI_SECRET}
        command: push
```

# Changelog

## 0.1.0
 - initial release
