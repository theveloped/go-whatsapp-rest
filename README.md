# Go WhatsApp Implementation in REST API

This is a fork from the [dimaskiddo/go-whatsapp-rest](https://github.com/dimaskiddo/go-whatsapp-rest) package. The addition of receiving text and image message using a webhook will be added to the original project.

## Generating RSA keys

To get started you will need an RSA private and public keypair. Get started by generating them using OpenSSL like so:
```
openssl genrsa -out mykey.pem 2048
openssl rsa -in mykey.pem -pubout > mykey.pub
```

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.
See deployment for notes on how to deploy the project on a live system.

### Prerequisites

Prequisites package:
* Go (Go Programming Language)
* Dep (Go Dependencies Management Tool)
* Make (Automated Execution using Makefile)

Optional package:
* GoReleaser (Go Automated Binaries Build)
* Docker (Application Containerization)

### Installing

Below is the instructions to make this codebase running:
* Create a Go Workspace directory and export it as the extended GOPATH directory
```
cd <your_go_workspace_directory>
export GOPATH=$GOPATH:"`pwd`"
```
* Under the Go Workspace directory create a source directory
```
mkdir -p src/github.com/dimaskiddo/go-whatsapp-rest
```
* Move to the created directory and pull codebase
```
cd src/github.com/dimaskiddo/go-whatsapp-rest
git clone -b master https://github.com/dimaskiddo/go-whatsapp-rest.git .
```
* Run following command to pull dependecies package
```
make ensure
```
* Until this step you already can run this code by using this command
```
make run
```

## Running The Tests

Currently the test is not ready yet :)

## Deployment

To build this code to binaries for distribution purposes you can run following command:
```
make compile
```
The build result will shown in build directory

## Built With

* [Go](https://golang.org/) - Go Programming Languange
* [Dep](https://github.com/golang/dep) - Go Dependency Management Tool
* [GoReleaser](https://github.com/goreleaser/goreleaser) - Go Automated Binaries Build
* [Make](https://www.gnu.org/software/make/) - GNU Make Automated Execution
* [Docker](https://www.docker.com/) - Application Containerization

## Authors

* **Dimas Restu Hidayanto** - *Initial Work* - [DimasKiddo](https://github.com/dimaskiddo)

See also the list of [contributors](https://github.com/dimaskiddo/go-whatsapp-rest/contributors) who participated in this project

## Annotation

You can seek more information for the make command parameters in the [Makefile](https://raw.githubusercontent.com/dimaskiddo/go-whatsapp-rest/master/Makefile)
