// newclient.go
package main

import (
    "os"
    "fmt"
    "log"
    "net"
    "net/rpc/jsonrpc"
)

/*
* Structure for the Arguments for Stock Trade
*/
type Args struct {
    StockSymbolAndPercentage string
    Budget float32
}

/*
* Structure for the Arguments for Portfolio
*/
type Tid struct {
    TradeId int
}

/*
* Structure for the Reply for Stock Trade
*/
type StockResult struct {
    TradeId int
    Stocks string
    UnvestedAmount float32
}

/*
* Structure for the Reply for Stock Portfolio
*/
type PortfolioResult struct {
    Stocks string
    CurrentMarketValue float32
    UnvestedAmount float32
}

/*
* Main Function
*/
func main() {
    // Initialize variables
    var StockSymbolAndPercentage string
    var Budget float32
    var TradeId, choice int

    // Print Choices to the users
    fmt.Printf("1. Buy new stocks\n2. View your portfolio\nEnter option (1/2):")
    fmt.Scanln(&choice)

    // Dial the connection
    client, err := net.Dial("tcp", "127.0.0.1:1234")
    if err != nil {
        log.Fatal("dialing:", err)
    }

    // Evaluate the choices
    if choice == 1 {

        // Stock Trading Choice
        fmt.Printf("Stocks: ")
        fmt.Scanln(&StockSymbolAndPercentage)
        fmt.Printf("Budget: $")
        fmt.Scanln(&Budget)

        // Synchronous call
        args := &Args{StockSymbolAndPercentage, Budget}
        var reply StockResult
        c := jsonrpc.NewClient(client)
        err = c.Call("Stocks.Trade", &args, &reply)
        if err != nil {
            log.Fatal("arith error:", err)
        }

        //  Print the response back to client
        fmt.Println(reply)

    } else if choice == 2 {

        // Stock Portfolio Choice
        fmt.Printf("TradeID: ")
        fmt.Scanln(&TradeId)

        // Synchronous call
        tid := &Tid{TradeId}
        var reply PortfolioResult
        c := jsonrpc.NewClient(client)
        err = c.Call("Stocks.Portfolio", &tid, &reply)
        if err != nil {
            log.Fatal("arith error:", err)
        }

        // Print the response
        fmt.Println(reply)

    } else {  // Wrong Choice
        fmt.Errorf("Wrong choice")
        os.Exit(1)
    }

}
