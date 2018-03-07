<!-- This document generated automatically -->
## Update project configuration.

### Update configuration

Update configurations by supplied command options.

```
$ ginger config [options]
```

| option    | description                                                                  |
|:---------:|:----------------------------------------------------------------------------:|
| --profile | Accout profile name. If empty, ginger uses `default` or environment variable |
| --region  | Region name to deploy                                                        |
| --bucket  | S3 bucket name                                                               |
| --hook    | Deploy hook command                                                          |


## Deploy all

Deploy all functions, resources, storage items.

```
$ ginger deploy all [options]
```

| option  | description                                                       |
|:-------:|:-------------------------------------------------------------------:|
| --stage | Stage name. If this option is supplied, create deployment to stage. |


## Deploy functions

Build and deploy lambda functions to AWS.

```
$ ginger deploy function [options]
```

| option | description                                                       |
|:------:|:-----------------------------------------------------------------:|
| --name | Function name. if this option didn't supply, deploy all functions |


## Deploy resources

Deploy API Gateway resources to AWS.

```
$ ginger deploy resource [options]
```

If resource has some integrations, create integration as well.

| option  | description                                                         |
|:-------:|:-------------------------------------------------------------------:|
| --stage | Stage name. If this option is supplied, create deployment to stage. |


## Deploy storage items

Deploy storage files to S3.

```
$ ginger deploy storage
```


## Create new function

Create new lambda function.

```
$ ginger function create [options]
```

| option  | description                                                                                              |
|:-------:|:--------------------------------------------------------------------------------------------------------:|
| --name  | Function name. If this option isn't supplied, ginger will ask it                                         |
| --event | Function event source. function template switches by this option. enable values are `s3` or `apigateway` |


## Delete function

Delete lambda function.

```
$ ginger function delete [options]
```

| option  | description              |
|:-------:|:------------------------:|
| --name  | [Required] function name |


## Invoke function

Invoke lambda function.

```
$ ginger function invoke [options]
```

| option  | description                                                                                           |
|:-------:|:-----------------------------------------------------------------------------------------------------:|
| --name  | [Required] function name                                                                              |
| --event | Passing event source data. data must be JSON format string, or can specify file name like `@filename` |


## List function

List registered lambda functions.

```
$ ginger function list
```


## Log function

Tailing log function output via CloudWatch Log.

```
$ ginger function log [options]
```

| option  | description                                                                                              |
|:-------:|:--------------------------------------------------------------------------------------------------------:|
| --name  | Function name. If this option isn't supplied, ginger will ask it                                         |


## Mount function

Create function integration to destination resource.

```
$ ginger function mount [options]
```

| option   | description                                                      |
|:--------:|:----------------------------------------------------------------:|
| --name   | Function name. If this option isn't supplied, ginger will ask it |
| --path   | Resource path. If this option isn't supplied, ginger will ask it |
| --method | Integration method                                               |


## Unmount function

Delete function integration to destination resource.

```
$ ginger function unmount [options]
```

| option   | description                                                      |
|:--------:|:----------------------------------------------------------------:|
| --path   | Resource path. If this option isn't supplied, ginger will ask it |
| --method | Integration method                                               |


## Initialize project

Initialize ginger project at current directory.

```
$ ginger init [options]
```

If you want to use (probably almost case yes) external Go package, we suggest you should put project directory under the `$GOPATH` to enable to detect vendor tree.

For example:

```
cd $GOATH/src/github.com/your/project
ginger init
```

The ginger init command will work as following:

- Create `Ginger.toml` file which is used for project configuration
- Create `functions` directory which is used for function management
- Create `stages` directory which is used for stage variable management
- Create `.ginger`  directory which is used for put dependency packags. Those packages will be loaded on deploy phase..
- Install dependency packages.

Note that the `Ginger.toml` is readable and configurable, but almost values are added or updated via subcommands.
So we don't recommend you change this file manually.

And, when initializing project, ginger asks two questions.

#### LambdaExecutionRole

When ginger deploys function to AWS Lambda, execution role is necessary.
So you should input lambda exection role to use as default. You can create role on AWS IAM.
See: https://docs.aws.amazon.com/lambda/latest/dg/with-s3-example-create-iam-role.html

Or, you can use specific role by each function by adding `Function.toml`.

#### S3BucketName

ginger uses S3 bucket name project director name as defaut. You can change this name.


## Install dependencies

Install dependency packages for build lambda function.

```
$ ginger install
```

This command is run automatically on initialize, but if you checkout project after initialize,
You can install dependency packages via this command.


## Create new scheduler

Create new cloudwatch scheduler .

```
$ ginger scheduler create [options]
```

| option  | description                                                                                              |
|:-------:|:--------------------------------------------------------------------------------------------------------:|
| --name  | Function name. If this option isn't supplied, ginger will ask it                                         |

After defined name, ginger want to input `expression`, you need to input CloudWatchEvent expression.
see: https://docs.aws.amazon.com/AmazonCloudWatch/latest/events/ScheduledEvents.html


## Delete scheduler

Delete CloudWatchEvent scheduler.

```
$ ginger scheduler delete [options]
```

| option  | description               |
|:-------:|:-------------------------:|
| --name  | [Required] scheduler name |


## List schedulers

List registered schedulers.

```
$ ginger scheduler list
```


## Attach scheduler to Lambda function

Relates scheduler to Lambda function.

```
$ ginger scheduler attach [options]
```

| option  | description                                                                                              |
|:-------:|:--------------------------------------------------------------------------------------------------------:|
| --name  | Scheduler name. If this option isn't supplied, ginger will ask it                                        |

Ginger will ask attach target function name by list UI.
