# running-man

RESTful JSON API for importing and aggregating road running results. 

### Installation

```sh
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
$ ./running-man -config ./config.json migrate-db

# start the http server
$ ./running-man -config ./config.json serve
```

### Importing Race Results

```sh
 curl -X POST http://localhost:8080/import -H "Accept: application/json" -d '{"raceUrl":"http://www.nlaa.ca/results/rr/2015/20150412flatout5k.php"}'
```

### List Races

```sh
 curl http://localhost:8080/races
```