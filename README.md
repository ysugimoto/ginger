# ginger - A Serverless framework for Go runtime

`ginger` is the framework that manages `AWS API Gateway` resources and `AWS Lambda` functions written in `go1.x` runtime.
It can creates `Serverless` architecture on `AWS` platform.

## Features

- Create / Delete `API Gateway` resources
- Create / Update / Delete `AWS Lambda` functions which works on `go1.x` runtime.
- Easily to make integration between `API Gateway` and `AWS Lambda`.

## Requirements

- Go (recommend latest version)
- AWS user who has some roles

## Installation

[url put here]

Move to your `$PATH` directory e.g. `/usr/local/bin`.

## Getting Started


### Setup

At the first, run the `ginger init` command on your project directory:


```
cd /path/to/project
ginger init
>>> some output...
```

If you want to use (probably almost case yes) external Go package, we suggest you put project directory under the `$GOPATH/src` to enable verndoring.

For example:

```
export GOPATH=$HOME/go
cd $HOME/go/src/ginger-project
ginger init
```

The `ginger init` command will work as following:

- Create `Ginger.toml`. It is a project configuration file
- Create `functions` directory. It is a function management directory
- Create `vendor` directory. It is a dependency vendor tree which will be loaded from go runtime.
- Install dependency packages.

### Project configuration

The `ginger` has three of project configurations. default values are deifned as following tables:

| Configuration Name  | Default Value | Description                                                                          |
|:-------------------:|:-------------:|:-------------------------------------------------------------------------------------|
| Profile             | (empty)       | Use profile name on AWS Request. If empty, ginger will use from environment variable |
| Region              | us-east-1     | AWS region which project use                                                         |
| LambdaExecutionRole | (empty)       | Set the AWS Lambda execution role                                                    |

Above configurations can change through a `config` subcommand:

```
ginger config --profile [Profile] --region [Region] --role [LambdaExecutionRole]`
```

Note that the `LambdaExecutionRole` is necessary to execute lambda function. Please make sure this value is set and role is valid.

And once you deployed some functions or apis, you __should not__ change the region because when region is changed, function will be created on different regions as same name.

### Create function

To create new function, run the `ginger create function --name [function name]` command.

`ginger` creates function structure under the `functions/` directory, and write out to configuration of `Ginger.toml`.

```
ginger function create --name example
```

You can find a `functions/example` directory which contains `main.go`. The `main.go` is a lambda function handler. The `github.com/aws/aws-lambda-go/lambda` is installed as default.

Of course you can install additional package with `go get` or other verndoring tools like `glide`, `dep`, ...

Note that `ginger function create` creates function only on your local. To work on `AWS Lambda`, you need to `deploy function`.

### Deploy function

After you modified a function, run `ginger deploy function` command to deploy to the `AWS Lambda`.

```
ginger deploy function
```

Command compiles function automatically, and archive to `zip`, and send to `AWS` to create on destination region.

Or, `ginger function deploy` is alias of this command so you can also use it to deploy function.

### Modify function

A Lambda function has a couple of settings:

| Name       | Default Value | Description                                                              |
|:----------:|:-------------:|:-------------------------------------------------------------------------|
| MemorySize | 128 (MB)      | Function memory size. this value must be above `128`, and multiple of 64 |
| Timeout    | 3 (Sec)       | Function timeout duration                                                |

In detail, see the [aws lambda documentation](https://docs.aws.amazon.com/lambda/latest/dg/limits.html)

To modify those values, run the `ginger function config` command with following options:

```
ginger function config --name [fucntion name] --memory [memory size] --timeout [timeout]
```

After that, the function configuration has been updated on your local, so you need to deploy to reflect those values.

### Invoke function

Once you deployed function to `AWS` by `ginger deploy functions` command, you can invoke the function via `AWS Lambda`:

```
ginger function invoke --name [function name] --event [event source json]
```

An `--event` option indicates event source for input of lambda function handler. `ginger` will pass the payload as this options:

- If option doesn't exists, pass as _empty payload_
- If option supplied as string, pass through
- If option starts with `@`, like `curl`, try to load the file and pass its content

After invocation end, the result print on your terminal.

To see in details, run the help command:

```
ginger function help
```

### Create Resource Endpoint

TODO: write

## License

MIT

## Author

ysugimoto (Yoshiaki Sugimoto)


