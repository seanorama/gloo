# End-to-end tests
This directory contains end-to-end tests that do not require Kubernetes

*Note: All commands should be run from the root directory of the Gloo repository*

## Background

This is the most common type of end-to-end test, since it is the quickest to set up and easiest to debug. However, it does not rely on Kubernetes, so if there is any Kubernetes behavior that needs to be tested, we recommend using the [kubernetes end-to-end tests](../kube2e) instead.

### How do the tests work?
1. Run the [Gloo controllers in goroutines](https://github.com/solo-io/gloo/blob/1f457f4ef5f32aedabc58ef164aeea92acbf481e/test/services/gateway.go#L109)
1. Run [Envoy](https://github.com/solo-io/gloo/blob/1f457f4ef5f32aedabc58ef164aeea92acbf481e/test/services/envoy.go#L237) either using a binary or docker container
1. Apply Gloo resources using [in-memory resource clients](https://github.com/solo-io/gloo/blob/1f457f4ef5f32aedabc58ef164aeea92acbf481e/test/services/gateway.go#L175)
1. Execute requests against the Envoy proxy and confirm the expected response. This validates that the Gloo resources have been picked up by the controllers, were been translated correctly into Envoy configuration, the configuration was sent to the Envoy proxy, and the proxy behaves appropriately.

## CI

These tests are run by [build-bot](https://github.com/solo-io/build-bot) as part of our CI pipeline.

### What if a test fails on a Pull Request?

Tests must account for the eventually consistent nature of Gloo Edge. If they do not wait for resources to be applied and processed correctly, they may flake. 

The best way to identify that a flake occurred is to run the test locally. We recommend [focusing the test](https://onsi.github.io/ginkgo/#focused-specs) to ensure that no other tests are causing an impact, and following the Ginkgo recommendations for [managing flaky tests](https://onsi.github.io/ginkgo/#repeating-spec-runs-and-managing-flaky-specs).

If a test failure is deemed to be a flake, we take the following steps:
1. Determine if there is a [GitHub issue](https://github.com/solo-io/gloo/labels/Type%3A%20CI%20Test%20Flake) tracking the existence of that test flake
1. Timebox an investigation into the flake. Flakes impact the developer experience, and we want to resolve them as soon as they are identified
1. If a solution can not be determined within a reasonable amount of time (1 hour), we create a GitHub issue to track it
1. If no issue exists, create one and include the `Type: CI Test Flake` label. If an issue already exists, add a comment with the logs for the failed run. We use comment frequency as a mechanism for determining frequency of flakes
1. Using the build-bot [comment directives](https://github.com/solo-io/build-bot#issue-comment-directives), retry the build, including a link to the GitHub issue tracking the flake in the comment on the Pull Request

## Local Development

### Setup

For these tests to run, we require Envoy be built in a docker container.

Refer to the [Envoyinit README](https://github.com/solo-io/gloo/blob/master/projects/envoyinit) for build instructions.


### Run Tests
The `run-tests` make target runs ginkgo with a set of useful flags. The following environment variables can be configured for this target:

| Name            | Default | Description |
| ---             |   ---   |    ---      |
| ENVOY_IMAGE_TAG | ""      | The tag of the gloo-envoy-wrapper-docker image built during setup |
| TEST_PKG        | ""      | The path to the package of the test suite you want to run  |
| WAIT_ON_FAIL    | 0       | Set to 1 to prevent Ginkgo from cleaning up the Gloo Edge installation in case of failure. Useful to exec into inspect resources created by the test. A command to resume the test run (and thus clean up resources) will be logged to the output.

Example:
```bash
ENVOY_IMAGE_TAG=solo-test-image TEST_PKG=./test/e2e/... make run-tests
```


### Debug Tests

#### Use WAIT_ON_FAIL
When Ginkgo encounters a [test failure](https://onsi.github.io/ginkgo/#mental-model-how-ginkgo-handles-failure) it will attempt to clean up relevant resources, which includes stopping the running instance of Envoy.

To avoid this clean up, run the test(s) with `WAIT_ON_FAIL=1`. When the test fails, it will halt execution, allowing you to inspect the state of the Envoy instance.

Once halted, use `docker ps` to determine the admin port for the Envoy instance, and follow the recommendations for [debugging Envoy](https://github.com/solo-io/gloo/tree/master/projects/envoyinit#debug), specifically the parts around interacting with the Administration interface.

## Additional Notes

### Notes on EC2 tests
*Note: these instructions are out of date, and require updating*

- set up your ec2 instance
  - download a simple echo app
  - make the app executable
  - run it in the background

```bash
wget https://mitch-solo-public.s3.amazonaws.com/echoapp2
chmod +x echoapp2
sudo ./echoapp2 --port 80 &
```

### Notes on AWS Lambda Tests (`test/e2e/aws_test.go`)

In addition to the configuration steps provided above, you will need to do the following to run the [AWS Lambda Tests](https://github.com/solo-io/gloo/blob/master/test/e2e/aws_test.go) locally:
  1. Obtain an AWS IAM User account that is part of the Solo.io organization
  2. Create an AWS access key
       - Sign into the AWS console with the account created during step 1
       - Hover over your username at the top right of the page. Click on "My Security Credentials"
       - In the section titled "AWS IAM credentials", click "Create access key" to create an acess key
       - Save the Access key ID and the associated secret key
  3. Install AWS CLI v2
       - You can check whether you have AWS CLI installed by running `aws --version`
       - Installation guides for various operating systems can be found [here](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2.html)
  4. Configure AWS credentials on your machine
       - You can do this by running `aws configure`
       - You will be asked to provide your Access Key ID and Secret Key from step 2, as well as the default region name and default output format
         - It is critical that you set the default region to `us-east-1`
       - This will create a credentials file at `~/.aws/credentials` on Linux or macOS, or at `C:\Users\USERNAME\.aws\credentials` on Windows. The tests read this file in order to interact with lambdas that have been created in the Solo.io organization
    