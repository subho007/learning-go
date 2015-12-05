package main

import (
	"fmt"
	"log"
  "net"
	"net/rpc/jsonrpc"
)

type Args struct {
	StockSymbolAndPercentage string
	Budget float32
}

type StockResult struct {
	tradeId int
	stocks string
	unvestedAmount float32
}

func main() {
	var input string
	var budget float32
	fmt.Printf("Stocks: ")
	fmt.Scanln(&input)
  fmt.Printf("budget: ")
	fmt.Scanln(&budget)
  //First DIAL
  client, err := net.Dial("tcp", "127.0.0.1:1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	// Synchronous call
	c := jsonrpc.NewClient(client)

	args := &Args{"GOOG:50%", 100.00}
	var reply StockResult
	var err1 = c.Call("Arith.Multiply", args, &reply)
	if err1 != nil {
		log.Fatal("arith error:", err1)
	}
	fmt.Println(reply.stocks)
	fmt.Println(reply.unvestedAmount)
	fmt.Println(reply.tradeId)
}
