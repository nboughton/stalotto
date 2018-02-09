# stalotto
Revisiting playing with Lotto data, this time as a CLI application

# Install
    go get github.com/nboughton/stalotto
    go install github.com/nboughton/stalotto

Run a db update after installation to pull all the historical data from the lotto website.

# Help
   stalotto -h

    Pull lotto results from web and present data derived from them
    
    Usage:
        stalotto [command]
    
    Available Commands:
        dip         Draw some random balls
        help        Help about any command
        records     Retrieve and print a record set
        update      Update or create the DB
    
    Flags:
          --db string   Set path to application db (default "/home/nick/.cache/stalotto/data.db")
      -h, --help        help for stalotto
    
    Use "stalotto [command] --help" for more information about a command.