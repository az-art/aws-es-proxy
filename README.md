# Amazon ElasticSearch proxy [![CircleCI](https://circleci.com/gh/az-art/aws-es-proxy.svg?style=shield)](https://circleci.com/gh/az-art/aws-es-proxy)

Amazon ElasticSearch proxy is a small web server which proxying HTTP requests between a service (app, curl) and Amazon ElasticSearch service. It will add sign header to all requests using latest "AWS Signature Version 4" and send back response to your client.

(as well for Kibana)

## Installation

### Download binary executable

**aws-es-proxy** has single executable binaries for Linux, Mac and Windows.

Download the latest [aws-es-proxy release](https://github.com/az-art/aws-es-proxy/releases/).

## Configuring Credentials

Before using **aws-es-proxy**, ensure that you've configured your AWS IAM user credentials. The best way to configure credentials on a development machine is to use the `~/.aws/credentials` file, which might look like:

```
[default]
aws_access_key_id = AKID1234567890
aws_secret_access_key = MY-SECRET-KEY
```

Alternatively, you can set the following environment variables:

```
export AWS_ACCESS_KEY_ID=AKID1234567890
export AWS_SECRET_ACCESS_KEY=MY-SECRET-KEY
```

**aws-es-proxy** also supports `IAM roles`. To use IAM roles, you need to modify your Amazon Elasticsearch access policy to allow access from that role. Below is an Amazon Elasticsearch `access policy` example allowing access from any EC2 instance with an IAM role called `ec2-aws-elasticsearch`.

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::012345678910:role/ec2-aws-elasticsearch"
      },
      "Action": "es:*",
      "Resource": "arn:aws:es:eu-west-1:012345678910:domain/test-es-domain/*"
    }
  ]
}
```



## Usage example:

```sh
./aws-es-proxy -e https://test-es-somerandomvalue.eu-west-1.es.amazonaws.com
Listening on 0.0.0.0:9200
```

*aws-es-proxy* listens on 127.0.0.1:9200 if no additional argument is provided. You can change the IP and Port passing the argument `-listen`

```sh
./aws-es-proxy -l :8080 -e ...
./aws-es-proxy -l 10.0.0.1:9200 -e ...
```

By default, *aws-es-proxy* will not display any message in the console. However, it has the ability to print requests being sent to Amazon Elasticsearch, and the duration it takes to receive the request back. This can be enabled using the option `-verbose`

```sh
./aws-es-proxy -v ...
Listening on 127.0.0.1:9200
2016/10/31 19:48:23  -> GET / 200 1.054s
2016/10/31 19:48:30  -> GET /_cat/indices?v 200 0.199s
2016/10/31 19:48:37  -> GET /_cat/shards?v 200 0.196s
2016/10/31 19:48:49  -> GET /_cat/allocation?v 200 0.179s
2016/10/31 19:49:10  -> PUT /my-test-index 200 0.347s
```

For a full list of available options, use `-h`:

```sh
./aws-es-proxy -h
Usage of ./aws-es-proxy:
  -e string
        Amazon ElasticSearch Endpoint (e.g: https://dummy-host.eu-west-1.es.amazonaws.com)
  -l string
        Local address listen on (default "127.0.0.1:9200")
  -p string
        Local TCP port listen to
  -logtofile
        Log user requests and ElasticSearch responses to files
  -nosignreqs
        Disable AWS Signature v4
  -pretty
        Prettify verbose and file output
  -v
        Print user requests
```

---