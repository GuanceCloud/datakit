# DataKit Control App (DCA)

DCA is a tool for DataKit management which runs on a web server.

## Build DCA web server

The web server is based on Go framework [Gin](https://gin-gonic.com/).

### Development

- #### start

```shell
$ npm start
```

### Build web image

Build image

```shell
$ sh build-dca-image.sh 0.0.1 # build image
```

Run a container

```shell
$ docker run -d -p 8000:80 --rm dca:0.0.1 # run container
```

Or run a container with custom environment variables

```shell
$ docker run -d -p 8000:80 --rm -e DCA_CONSOLE_WEB_URL="https://console.guance.com" -e DCA_CONSOLE_API_URL="https://console-api.guance.com" -e DCA_LOG_LEVEL="INFO" dca:0.0.1
```
- `DCA_CONSOLE_API_URL`: console API host, default `https://console-api.guance.com`
- `DCA_CONSOLE_WEB_URL`: console web host, default `https://console.guance.com`
- `DCA_LOG_LEVEL`: log level, NONE | DEBUG | INFO | WARN | ERROR
- `DCA_LOG_ENABLE_STDOUT`: whether to log to stdout
