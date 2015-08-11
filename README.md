# distelli

[![wercker status](https://app.wercker.com/status/8a086b544ada1f28962252b35b4de1f3/m/master "wercker status")](https://app.wercker.com/project/bykey/8a086b544ada1f28962252b35b4de1f3)

Downloads the [Distelli CLI](https://www.distelli.com/docs/distelli-cli-reference) and runs a command.

# Best practices

This step allows a choice on the spectrum of flexibility and guard rails,
provided by the `branches` configuration key. If this key is set, then the
distelli step will be skipped. This allows a project to define a set of
branches from which deploys should be allowed.

Setting this value via an environment variable in Wercker allows for maximum
flexibility. With the set-up, different deploy targets can be allow to deploy
from different branches, or the set of branches can be temporarily changed
with out requiring an update to the wercker.yml in source control. See the
example below for an illustration.

## Options.

### required

* `command` - command to run. Currently only "supports" `push` and `deploy`.
* `accessKey` - the Distelli access key (token) for using the CLI
* `secretKey` - the Distelli secret key for using the CLI

### optional

* `branches` - a whilelist of branches to allow commands on. If unset, all branches are allowed.
* `manifest` - the distelli manifest file. Required for `push` and `deploy` commands.
* `releaseFilename` - a file name to track Distelli release id between `push` and `deploy` commands.
* `wait` - (do not) wait for a distelli deploy to finish before proceeding. Only supported for the `deploy` command.

# Example

Push a build to Distelli. The build will have a commit message of the form `wercker:${WERCKER_BUILD_ID}`.
This form is necessary to locate the pushed bundle for later deployment.

``` yaml
build:
  steps:
    - distelli:
        branches: ${ALLOWED_BRANCHES}
        accessKey: ${DISTELLI_TOKEN}
        secretKey: ${DISTELLI_SECRET}
        command: push
```

# Changelog

## 0.2.1
 - trap exit status

## 0.2.0
 - golang rewrite

## 0.1.0
 - initial release
