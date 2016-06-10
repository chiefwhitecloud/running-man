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
 curl -X POST http://localhost/import -H "Content-Type: application/json" -d '{"raceUrl":"http://www.nlaa.ca/results/rr/2015/20150412flatout5k.php"}'
```

If successful returns a status of 201 and the race json structure

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
