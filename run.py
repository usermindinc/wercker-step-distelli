#!/bin/env python

from __future__ import print_function
import csv
import os
import subprocess
import sys
import yaml

distelli = os.path.join(os.getenv("WERCKER_STEP_ROOT"), "DistelliCLI", "bin", "distelli")
cache_dir = os.getenv("WERCKER_CACHE_DIR")
git_branch = os.getenv("WERCKER_GIT_BRANCH")
git_commit = os.getenv("WERCKER_GIT_COMMIT")
output_dir = os.getenv("WERCKER_OUTPUT_DIR")
temp_dir = os.getenv("WERCKER_STEP_TEMP")


def message(text):
    print(text, file=sys.stderr)


def info(text):
    message(text)


def fail(text):
    message(text)
    exit(1)


def check_branches():
    branches = os.getenv("WERCKER_DISTELLI_BRANCHES")

    if branches is None:
        return

    for branch in branches.split(","):
        if branch == git_branch:
            return

    info("Current branch %s not in permitted set %s, skipping distelli step." % (git_branch, branches))
    exit(0)


def check_manifest():
    manifest = os.getenv("WERCKER_DISTELLI_MANIFEST")
    if manifest is None:
        fail("manifest must be set")

    if not os.path.exists(manifest):
        fail("manifest file %s not found" % manifest)

    return os.path.split(manifest)


def check_credentials():
    access_key = os.getenv("WERCKER_DISTELLI_ACCESSKEY")
    secret_key = os.getenv("WERCKER_DISTELLI_SECRETKEY")

    if not access_key or not secret_key:
        fail("Access key and secret key are required.")

    os.putenv("DISTELLI_TOKEN", access_key)
    os.putenv("DISTELLI_SECRET", secret_key)


def locate_app_name():
    app = os.getenv("WERCKER_DISTELLI_APPLICATION")

    if not app:
        (dirname, basename) = check_manifest()
        with open(os.path.join(dirname, basename), 'r') as stream:
            doc = yaml.load(stream)
            app = next(iter(doc))

        if not app:
            fail("Could not locate app name from manifest")

    return app


def locate_build_id():
    build_id = os.getenv("WERCKER_BUILD_ID")
    build_filename = os.path.join(cache_dir, "usermind-build-id.txt")

    if not build_id:
        with open(build_filename, 'r') as build_file:
            build_id = build_file.readline()
            os.putenv("WERCKER_BUILD_ID", build_id)
    else:
        message("Saving build id %s to %s" % (build_id, build_filename))
        with open(build_filename, 'w') as build_file:
            build_file.writelines([build_id, ''])

    return build_id


def locate_release_id():
    release_id = os.getenv("WERCKER_DISTELLI_RELEASE")
    build_id = locate_build_id()

    if not release_id:
        # Nothing was specified, so we need to query distelli and look for the release
        app = locate_app_name()

        output = invoke("list releases -n %s -f csv" % app, capture=True)
        reader = csv.reader(output.splitlines())
        for row in reader:
            description = row[3]
            if description == "wercker:%s" % build_id:
                release_id = row[1]
                # Releases are listed in creation order. In the case of failures,
                # it's possible multiple releases are created for the same build
                # id. Continue iterating in order to find the most recent release.

        if not release_id:
            fail("Unable to locate release for build %s in app %s" % (build_id, app))

    return release_id


def invoke(cmd, capture = False):
    (dirname, basename) = check_manifest()

    # Distelli 1.88 assumes manifest is in CWD
    old_cwd = os.getcwd()
    os.chdir(dirname)

    # Wercker checks us out to a commit, not a branch name (sensible, since the
    # branch may have moved on). Distelli doesn't handle this well. We won't have
    # any local branches (except master), so create one with an appropriate name.

    # Checkout the commit to ensure the branch is not current
    os.system("git checkout -q %s" % git_commit)

    # Force update the branch name
    os.system("git branch -f %s %s" % (git_branch, git_commit))

    # Switch to the branch
    os.system("git checkout -q %s" % git_branch)
    output = None

    try:
        with open(os.devnull) as fnull:
            if capture:
                output = subprocess.check_output("%s %s" % (distelli, cmd), stdin=fnull, shell=True)
            else:
                subprocess.check_call("%s %s" % (distelli, cmd), stdin=fnull, shell=True)
    except subprocess.CalledProcessError, ex:
        raise fail("Error executing distelli %s\n%s\n%s" % (cmd, ex.message, ex.output))

    os.chdir(old_cwd)
    return output


def push():
    (dirname, basename) = check_manifest()

    invoke("push -f %s -m wercker:%s" % (basename, locate_build_id()))


def deploy():
    args = []

    environment = os.getenv("WERCKER_DISTELLI_ENVIRONMENT")
    host = os.getenv("WERCKER_DISTELLI_HOST")

    if environment:
        if host:
            fail("Both environment and host are set")
        args.extend(["-e", environment])
    elif host:
        args.extend(["-h", host])
    else:
        fail("Either environment or host must be set")

    release_id = locate_release_id()
    (dirname, basename) = check_manifest()

    args.extend(["-y", "-f", basename, "-r", release_id, "-m", "wercker:%s" % locate_build_id()])

    cmd = "deploy %s" % " ".join(args)
    output = invoke(cmd, capture=True)
    if "Deployment Failed" in output:
        fail(output)


def main():
    os.system("%s version" % distelli)

    check_branches()
    check_credentials()

    command = os.getenv("WERCKER_DISTELLI_COMMAND")

    if command is None:
        fail("command must be set")

    elif command == "push":
        push()

    elif command == "deploy":
        deploy()

    else:
        fail("unknown command: %s" % command)

if __name__ == "__main__":
    main()

