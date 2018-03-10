# ginger - Serverless framework for Go runtime

`ginger` is the framework manages `Serverless` architecture for Go runtime.

## Features

`ginger` manages following AWS services:

- `API Gateway` endpoints
- `S3` storage files
- `Cloudwatch` schedule events and Lambda function logger
- `Lambda` __only for `go1.x` runtime__

## Requirements

- Go (we recommend latest version)
- AWS user who has above resource permissions

## Installation

You can download prebuilt binary at [Release](https://github.com/ysugimoto/ginger/releases).
But although you need to install `Go` for compile Lambda function.

To see a general usage, run `ginger help` command.

## Getting Started

### Setup

Run the `ginger init` command at your project directory:


```
cd /path/to/project
ginger init
>>> some output...
```

If you want to use (probably almost case yes) external Go package, we suggest you should put project directory under the `$GOPATH/src` to enable to detect vendor tree.

For example:

```
export GOPATH=$HOME/go
cd $HOME/go/src/ginger-project
ginger init
```

`ginger` wants to input `Lambda execution role` and `S3 storage name`, you should input suitable value.

### Create function

To create new function, run the `ginger fucntion create --name [function name]` command.

`ginger` creates function structure under the `functions/` directory, and write out to configuration of `Ginger.toml`.

```
ginger function create --name example
```

You can find a `functions/example` directory which contains `main.go` and `Function.toml`.
The `main.go` is a Lambda function handler. The `github.com/aws/aws-lambda-go/lambda` is installed as default.
On the other hand, the `Function.toml` is setting file of Lambda function e.g. memory limit, timeout, and so on.

Of course you can install additional package with `go get` or other verndoring tools like `glide`, `dep`, ...

Note that `ginger function create` creates function only on your local. To work on `AWS Lambda`, you need to `deploy function`.

### Deploy function

After you modified a function, run `ginger deploy function` command to deploy to the `AWS Lambda`.

```
ginger deploy function (--name [destination function])
```

`ginger` compiles function automatically and archive to `zip`, finally send to `AWS` to create on destination region.

Or `ginger function deploy` is alias of this command, so you can also use it to deploy function.

### Invoke function

Once you deployed function to `AWS`, you can invoke the function via `AWS Lambda`:

```
ginger function invoke --name [function name] --event [event source json]
```

An `--event` option indicates event source for input of lambda function handler. `ginger` gets the payload as following options and pass to the function input:

- If option doesn't exists, pass as _empty payload_
- If option supplied as string, pass as it is
- If option starts with `@`, like `curl`, ginger tries to load the file and pass its content

After invocation end, the result print on your terminal.

To see in details, run the help command:

```
ginger function help
```

### Create Resource Endpoint

To create API endpoint, run the `ginger resource create --path [endpoint path]` command.

`ginger` creates endpoint on `Ginger.toml`.

```
ginger resource create --path /foo/bar
```

Note that `ginger resource create` creates endpoint info only on your local. To work on `AWS API Gateway`, you need to `deploy api`.

### Deploy api

After you created endpoint, run `ginger deploy resource` command to deploy to the `AWS API Gateway`.

```
ginger deploy resource --stage [target stage] (--path [destination path])
```

Command creates resouce which we need, and also create root `REST API` if you haven't create it.

if `--stage` option id supplied, ginger tries to create deployment to target stage. Otherwise, only create resources.

Note that the `AWS API Gateway` manages endpoints as `pathpart`, it is part of segment, so we need to create recursively by each segment.
But you don't need to care about it because `ginger` creates and manages sub-path automatically and save on `Ginger.toml`.

In detail, see [AWS API Gateway documentation](https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-method-settings-method-request.html).

Or `ginger resource deploy` is alias of this command, so you can also use it to deploy resources.

### Setup Lambda Integration

The `AWS API Gateway` supports [Lambda Proxy Integration](https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-create-api-as-simple-proxy-for-lambda.html), and `ginger` can manage its feature.

To set up it, run `ginger function mount` command with function name and endpoint option:

```
ginger function mount
```

Then, `ginger` asks target function and endpoint by choosing list. After select both and deploy api to AWS, proxy integration creates automatically.
Let's access to API Gateway URL!

### Invoke endpoint

In default, the `API Gateway` endpoint is complicated a little. So you can invoke HTTP request through `ginger resource invoke` command with `--stage` option to determine invoke stage

```
ginger resource invoke --stage [stage name]
```

`ginger` asks path input, make request URI and send HTTP request, and outputs response headers and body.

## API Doc

See [Command API document](https://github.com/ysugimoto/ginger/blob/master/docs/command.md)

## Examples

Now writing...

## Development

Checkout this project and build locally:

```
cd $GOPATH
go get github.com/ysugimoto/ginger
cd src/github.com/ysugimoto/ginger
make
```

On `make` command builds with `debug flag`. This flag dumps stacktrace on error and all AWS SDK requests and responses.
It will help you how command processed.

We welcome your feedbacks and PRs :-)

## License

MIT

## Author

ysugimoto (Yoshiaki Sugimoto)


