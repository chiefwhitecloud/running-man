# running-man

RESTful JSON API for importing and aggregating road running results.

### Installation

```sh



# set environment variables

DATABASE_URL: Database connection string
PORT : Port number the service will run on
ASSET_PATH :  JS and CSS location

# create the database configured in `config.json`
$ mysql -u root -p -e "CREATE DATABASE RunningMan;"

# create the database configured in `test.json`
$ mysql -u root -p -e "CREATE DATABASE RunningManTest;"
```

```sh
$ go get github.com/chiefwhitecloud/running-man
$ cd $GOPATH/src/github.com/chiefwhitecloud/running-man
$ go build
$ go test ./...

# add the tables
$ ./running-man migrate-db

# start the http server
$ ./running-man serve
```


### Importing Race Results

```sh
 curl -X POST http://localhost/import
    -H "Content-Type: application/json"
    -d '{"raceUrl":"http://www.nlaa.ca/results/rr/2015/20150412flatout5k.php"}'
```

If the request is accepted, it will return a Status of 202 and a Location header containing the path to the import task resource.

```
Status: 202
Location: http://localhost/import/task/1
```

HTTP Get on the import task resource url will return a 200 if the task is pending.

When the task is successfully completed, the import task url will return a Status 303 with a redirect header containing the race resource url.

HTTP Get on the race resouce url returns a 200 Status and the following body.

```
  {
    "id":1,
    "name":"Boston Pizza Flat Out 5 km Road Race",
    "self":"http://localhost/feed/race/1",
    "results":"http://localhost/feed/race/1/results",
    "date":"2015-04-12"
  }
```

### List Races

```sh
 curl http://localhost/feed/races
```
