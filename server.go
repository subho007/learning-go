package main

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"net/rpc"
	"strconv"
	"os"
	"io/ioutil"
	"encoding/json"
	"net/rpc/jsonrpc"

)
type Args struct {
	stockSymbolAndPercentage string
	budget float32
}
type StockResult struct {
	tradeId int
	stocks string
	unvestedAmount float32
}

type Arith struct {}


func (t *Arith) Multiply(args *Args, reply *StockResult) error {
		stockAndPercent := args.stockSymbolAndPercentage
		stocksBudget := args.budget
		stockDetails := ""
    fmt.Println(args)
		if stockAndPercent =="" {
			fmt.Println("Please enter stocks")
		}

		if(strings.Contains(stockAndPercent,",")) {
			fmt.Println("todo")
		} else {
			companyStocks := strings.Split(stockAndPercent,":")
			stockSymbol := companyStocks[0]
			stockPercent := companyStocks[1]
			stringPercent := strings.Trim(stockPercent,"%")
			percentFloat64,_ := strconv.ParseFloat(stringPercent,64)
			percentFloat32 := float32(percentFloat64)

			stockPrice := FetchStockPrice(stockSymbol)
			stockPriceFloat64,_ := strconv.ParseFloat(stockPrice,64)
			stockPriceFloat32 := float32(stockPriceFloat64)

			availableAmount := (stocksBudget) * (percentFloat32/100)
			noOfSharesFloat := availableAmount/stockPriceFloat32
			noOfSharesInt := int(noOfSharesFloat)
			noOfSharesFloat32 := float32(noOfSharesInt)
			fmt.Println("Sakshi",companyStocks)

			amountRemaining := stocksBudget - (noOfSharesFloat32*stockPriceFloat32)

			noOfSharesInString := strconv.Itoa(noOfSharesInt)

			stockDetails = stockSymbol+":"+noOfSharesInString+":"+stockPrice
			reply.unvestedAmount = amountRemaining

		}
		reply.stocks= stockDetails
		reply.tradeId = 1
    return nil
}

func FetchStockPrice(stockSymbol string) string {
	req, err := http.Get("http://finance.yahoo.com/webservice/v1/symbols/"+stockSymbol+"/quote?format=json")
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	}
	defer req.Body.Close()
	contents, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	}
	result:= []byte(contents)
	var m(map[string]interface{})
	err = json.Unmarshal(result, &m)
	if err!=nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	}

	queryResponse:= m["query"].(map[string]interface{})
	resultResponse := queryResponse["results"].(map[string]interface{})
	quoteResponse := resultResponse["quote"].(map[string]interface{})
	shareAmount := quoteResponse["LastTradePriceOnly"].(string)
	return shareAmount

}

func main() {
	  fmt.Printf("I am here")
    arith := new(Arith)
    rpc.Register(arith)

    tcpAddr, err := net.ResolveTCPAddr("tcp", ":1234")
    checkError(err)

    listener, err := net.ListenTCP("tcp", tcpAddr)
    checkError(err)

    for {
        conn, err := listener.Accept()
        if err != nil {
            continue
        }
        jsonrpc.ServeConn(conn)
    }
}

func checkError(err error) {
    if err != nil {
        fmt.Println("Fatal error ", err.Error())
        os.Exit(1)
    }
}
