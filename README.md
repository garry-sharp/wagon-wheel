# Wagon Wheel

An attempt at an assessment I saw sent out recently. The goal is to create a CLI that gets price feeds from a cryptocurrency API and store them to a DB. Part 2 of the task is the write them to a very simple web page.

## Design

### CLI

The CLI was build using cobra with golang. Perhaps overboard for the needs of this task since subcommands aren't needed. Flags are passes which kick off a process to gather prices.

### Fetching Prices

Prices are loaded into a `Price` struct which uses gorm (a golang ORM) to turn that into a DB object and write/save to the DB. Cheating a little bit but since the data we're dealing with in this example, writing tables, queries and update/insert SQL statements seems excessive. For this reason too I decided to use sqlite and store the data in a file, that way I can avoid using docker.

### Endpoints and Web UI

The site also renders a very simple HTML page which uses a few libraries to cut corners (toastr, bootstrap, jQuery etc). The page can hit 2 "normal" http endpoints, to start and stop the server. Starting the server does not refresh the operation in progress but rather kicks off extra processes. This would allow you to say make 2 requests to get different quotes for different assets. Stop stops everything.

The 3rd endpoint is an "upgraded" endpoint and writes data to the page using web sockets. The HTTP controller basically sits in an infinite for loop and reads from a channel the prices get written to. Many thanks to gorilla/websocket for the library, made it quite painless.

### Logging

This took me longer than I wanted, I definitely over-engineered this part. I came up with an idea of using UNIX `chmod` style bitmaps to determine the logging type. With 1 (1st bit) being the flag for stdout logging, 2 (2nd bit) being the flag for file logging and 4 (3rd bit) being the flag for kakfa logging. So an input of 5 would give you kakfa and stdout, 3 would give you stdout and file, 7 would give you everything and so forth. This maybe isn't the best solution to this particular exercise but it could be useful if you extended the application to cover lots of different log scenarios.

If you choose to use kakfa or file logging you need to pass additional parameters over the CLI. For file logging (just like the SQLite DB filename) a default file is created if not specified. If a number `x` where `x & 4 == 1` is provided (the 3rd bit in the number is on) then a kakfa URL __must__ be provided.

Purely for demo purposes a `kakfa_docker-compose.yml` file has been added for convenience. In reality it's highly unlikely a kakfa service would be hosted on the same machine as the main application code. To launch the kakfa instance you can use `docker-compose -f kafta_docker-compose.yml up -d`. To kill it simply exit the application or `docker-compose down -f kafta_docker-compose.yml` to be sure. When up a local kafta URL can be passed via the main app CLI with `--kafkaurl https://localhost:9092`

Also fair warning, I'm probably not doing the kafta stuff correctly, this is my first time using it and I am sort of learning on the job.

## Errors

Pfft, plenty, I didn't spend loads of time on this and there are for sure areas where I could handle things better, unmarhsalling json, parsing strings to ints etc. As it stands if a pair doesn't exist it simply doesn't return any values, the being said however, the code does still analyse it as it is saved in a list of assets which are looped through. In reality it's not a big deal but with LOTS of usage it could be a performance issue. 

Again in "real" software I'd be a lot more careful about how I use this stuff.

## Build

This application was built using golang 1.18. Build should be very simple, first make sure you have go 1.18. You can download for your machine [here](https://go.dev/dl), then a case of `go build .` followed by `./assessment --help` will give you all the instructions you need. The web server and traditional CLI price logger are both launched using the CLI. It also uses kafka for logging, you'll need to pass it an endpoint for the logs to work