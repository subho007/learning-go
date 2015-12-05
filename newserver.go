// newserver.go
package main

import (
    "fmt"
    "log"
    "strings"
    "strconv"
    "net"
    "net/http"
    "net/rpc"
    "net/rpc/jsonrpc"
    "io/ioutil"
    "encoding/json"
)

/*
* Structure for the Arguments for Portfolio coming from the Client
*/
type Tid struct {
    TradeId int
}

/*
* Structure for the Arguments for Trade coming from the Client
*/
type Args struct {
    StockSymbolAndPercentage string
    Budget float32
}

/*
* Structure to send back the portfolio result to the client
*/
type PortfolioResult struct {
    Stocks string
    CurrentMarketValue float32
    UnvestedAmount float32
}

/*
* Structure to send back the trade result to the client
*/
type StockResult struct {
    TradeId int
    Stocks string
    UnvestedAmount float32
}

/*
* Collection of trade results to store
*/
type StockResults struct {
    Items []StockResult
}

/*
* Helper function to add Items to the trade results
*/
func (query *StockResults) AddItem(item *StockResult) []StockResult {
    query.Items = append(query.Items, *item)
    return query.Items
}

/*
* Global variable to store list of trade results to refer from the portfolio
*/
var Result StockResults

/*
* The JSON Structure of YAHOO Finance JSON API
*/
type StockData struct {

    List struct{

        Meta struct{
            Count int   `json:"count"`
            Start int   `json:"start"`
            Type string `json:"type"`
        } `json:"meta"`

        Resources []struct{

            Resource struct{

                Classname string `json:"classname"`

                Fields struct{
                    Name string    `json:"name"`
                    Price string   `json:"price"`
                    Symbol string  `json:"symbol"`
                    Ts string      `json:"ts"`
                    Type string    `json:"type"`
                    UTCtime string `json:"utctime"`
                    Volume string  `json:"volume"`
                }`json:"fields"`

            }`json:"resource"`

        }`json:"resources"`

    }`json:"list"`
}

/*
* Structure for storing each stock query items during trade
*/
type StockQueryItem struct {
    StockName string
    StockPercentage float64
    StockBudget float64
}

/*
* Structure to store the information of stocks during trade
*/
type StockValue struct {
    StockName string
    StockValue float64
    StockQuantity int
}

/*
* Collection of stocks during trading
*/
type StockValues struct {
    Items []StockValue
}

/*
* Collection of query which the user has asked
*/
type StockQuery struct {
    Items []StockQueryItem
}

/*
* Helper function to add Items to the []structure
*/
func (query *StockValues) AddItem(item StockValue) []StockValue {
    query.Items = append(query.Items, item)
    return query.Items
}

/*
* Structure for the stock RPC
*/
type Stocks struct{}

/*
* Helper function to add Items to the stockQuery structure
*/
func (query *StockQuery) AddItem(item StockQueryItem) []StockQueryItem {
    query.Items = append(query.Items, item)
    return query.Items
}

/*
* This gets called when RPC is invoked by Stocks.Portfolio
*/
func (t *Stocks) Portfolio(args *Tid, reply *PortfolioResult) error {
    tradeid := args.TradeId
    stockdata := &StockData{}
    // Checking if the tradeId exists or not
    if !check_tradeid_exist(tradeid) {
        return nil
    }
    stocks := Result.Items[tradeid - 1].Stocks  // See logic of generating tradeid
    symbols := get_portfolio_symbols(stocks)  // gets the symbols in terms of GOOGL,APPL
    var err = json.Unmarshal(fetch_stock_yahoo(symbols), stockdata)  // Fill in the structure of the YAHOO Finance Webservice API
    if err != nil {
        fmt.Errorf("Stocks can not be parsed from Yahoo API")
	}

    reply.Stocks = format_reply_portfoliostocks(stocks, stockdata)
    reply.CurrentMarketValue = format_reply_currentmarketvalue(stocks, stockdata)
    reply.UnvestedAmount = Result.Items[tradeid - 1].UnvestedAmount
    // DEBUG Println
    fmt.Println(Result.Items[tradeid - 1])
    return nil
}

/*
* Function to format the reply of the PortFolio Stocks
* @params stocks: String which holds values like `GOOGL:0:$656.99,APPL:5:$96.11`
* @params stockdata: Stucture of the YAHOO API
* @return: Formatted string like 'GOOGL:0:+$656.99','APPL:5:+$96.11'
*/
func format_reply_portfoliostocks(stocks string, stockdata *StockData) string {
    var value string
    stock := strings.Split(stocks, ",")
    for i :=0; i < len(stockdata.List.Resources); i++ {
        StockName := stockdata.List.Resources[i].Resource.Fields.Symbol
        for j :=0; j < len(stock); j++ {
            if !strings.EqualFold(StockName, strings.Split(stock[j], ":")[0]) {
                continue
            }
            CurrentMarketValue, _ := strconv.ParseFloat(stockdata.List.Resources[i].Resource.Fields.Price, 64)
            StockPrice, _ := strconv.ParseFloat(strings.Split(stock[j], ":")[2], 64)
            quantity:= strings.Split(stock[j], ":")[1]
            value = value + ",'"+ StockName + ":" + quantity
            if CurrentMarketValue > StockPrice {
                value = value + ":+$"
            } else if CurrentMarketValue < StockPrice {
                value = value + ":-$"
            } else {
                value = value + ":$"
            }
            value = value + strconv.FormatFloat(CurrentMarketValue, 'f', 2, 64) + "'"
        }
    }
    return strings.Trim(value, ",")
}

/*
* Function to format the reply of the PortFolio currentmarketvalue
* @params stocks: String which holds values like `GOOGL:0:$656.99,APPL:5:$96.11`
* @params stockdata: Stucture of the YAHOO API
* @return: Formatted float32 of the currentmarketvalue which is the sum of all the stock prices * the quantity
*/
func format_reply_currentmarketvalue(stocks string, stockdata *StockData) float32 {
    var value float64
    stock := strings.Split(stocks, ",")
    for i :=0; i < len(stockdata.List.Resources); i++ {
        StockName := stockdata.List.Resources[i].Resource.Fields.Symbol
        for j :=0; j < len(stock); j++ {
            if !strings.EqualFold(StockName, strings.Split(stock[j], ":")[0]) {
                continue
            }
            Price, _ := strconv.ParseFloat(stockdata.List.Resources[i].Resource.Fields.Price, 64)
            quantity, _ := strconv.ParseFloat(strings.Split(stock[j], ":")[1], 64)
            value = value + (Price * quantity)
        }
    }
    return float32(value)
}

/*
* Function to check if the tradeid exists or not in the portfolio call
* @params TradeId: Integer
* @rtype Boolean
*/
func check_tradeid_exist(TradeId int) bool {
    for _, each := range Result.Items {
        if each.TradeId == TradeId {
            return true
        }
    }
    return false
}

/*
* Function to get all the symbols of the stocks from a stocks
* @params stocks: String which holds values like `GOOGL:0:$656.99,APPL:5:$96.11`
* @rtype: returns GOOGL,APPL
*/
func get_portfolio_symbols(stocks string) string {
    var symbols string
    for _, each := range strings.Split(stocks, ",") {
        symbol := strings.Split(each, ":")[0]
        symbols = symbols + "," + symbol
    }
    return strings.Trim(symbols, ",")
}


/*
* This gets called when RPC is invoked by Stocks.Trade
*/
func (t *Stocks) Trade(args *Args, reply *StockResult) error {
    fmt.Println(args)
    fmt.Println(args.StockSymbolAndPercentage)
    var StockQueries StockQuery
    StockQueries = get_stock_query(strings.TrimSpace(args.StockSymbolAndPercentage), args.Budget)
    stockdata := &StockData{}
    stocksymbol := get_stock_symbol(StockQueries)
    var err = json.Unmarshal(fetch_stock_yahoo(stocksymbol), stockdata)
    if err != nil {
        fmt.Errorf("Stocks can not be parsed from Yahoo API")
	}
    if stockdata.List.Meta.Count != len(StockQueries.Items) {
        panic("One of the symbols are wrong, please try again")
    }
    stockvalues := compute_buy_stocks(StockQueries, *stockdata)
    TradeId := len(Result.Items) + 1
    reply.TradeId = TradeId
    reply.Stocks = format_reply_stocks(stockvalues)
    reply.UnvestedAmount = args.Budget - format_reply_amount(stockvalues)
    Result.AddItem(reply)
    fmt.Println(stockvalues.Items)
    return nil
}

/*
* Function to return a string formatted for reply during stock trading
* @params StockValues
* @rtype: Returns string like `GOOGL:0:$656.99,APPL:5:$96.11`
*/
func format_reply_stocks(stockvalues StockValues) string {
    var stocks string
    for _, each := range stockvalues.Items {
        stocks = stocks + each.StockName + ":" + strconv.Itoa(each.StockQuantity) + ":$" + strconv.FormatFloat(each.StockValue, 'f', 2, 64) + ","
    }
    return strings.Trim(stocks, ",")
}

/*
* Function to return the total amount invested in stock during trading
*/
func format_reply_amount(stockvalues StockValues) float32 {
    var stocks float32
    for _, each := range stockvalues.Items {
        stocks = stocks + float32(each.StockValue)
    }
    return stocks
}

/*
* Function to compute the purchase of stock properly and store it in reply struc
*/
func compute_buy_stocks(StockQueries StockQuery, stockdata StockData) StockValues {
    stockvalues := StockValues{[]StockValue{}}
    for i := 0; i < len(StockQueries.Items); i++ {
        StockName := StockQueries.Items[i].StockName
        StockBudget := StockQueries.Items[i].StockBudget
        for j :=0; j < len(stockdata.List.Resources); j++ {
            Price, _ := strconv.ParseFloat(stockdata.List.Resources[i].Resource.Fields.Price, 64)
            Symbol := stockdata.List.Resources[i].Resource.Fields.Symbol
            if !strings.EqualFold(StockName, Symbol) {
                continue
            }
            quantity := int(StockBudget/Price)
            stockvalue := StockValue{StockName: Symbol, StockValue: Price, StockQuantity: quantity}
            stockvalues.AddItem(stockvalue)
            break;
        }
	}
    return stockvalues
}


/*
* Function to fetch the stock pricing and details from Yahoo
*  @params: StockSymbol - a string of the required Stock symbols, ex: GOOGL,APPL
*  @return: []byte of the JSON RAW Response
*/
func fetch_stock_yahoo(StockSymbol string) []byte {
    url := fmt.Sprintf("http://finance.yahoo.com/webservice/v1/symbols/%s/quote?format=json", StockSymbol)
    // Do a GET Request to get the JSON formatted file from the server
    req, err := http.Get(url)
    if err != nil {
        panic(err.Error())
    }
    // Get the BODY of the Request
    body, err := ioutil.ReadAll(req.Body)
    if err != nil {
        panic(err.Error())
    }
    defer req.Body.Close()  // Wait for the request to be completed [Blocking Call]
    return body
}

/*
* Function to return all the STOCK symbol in this manner APPL,GOOGL
*/
func get_stock_symbol(stockquery StockQuery) string {
    var all_symbols, symbol string
    for _, each := range stockquery.Items {
        symbol = each.StockName
        all_symbols = all_symbols + symbol + ","
    }
    return strings.Trim(all_symbols,",")
}

/*
*  Function to compute/verify and to aggregate all the symbols of the companies
*/
func get_stock_query(StockSymbolAndPercentage string, Budget float32) StockQuery {
    var Percent, _StockBudget float64
    var _Percent string
    items := []StockQueryItem{}
    query := StockQuery{items}
    for _, each := range strings.Split(StockSymbolAndPercentage, ",") {
        var Name = strings.Split(each, ":")[0]
        if len(Name) < 2 {
            continue
        }
        _Percent = strings.Split(each, ":")[1]
        Percent, _ = strconv.ParseFloat(strings.Trim(_Percent, "%"), 64)
        _StockBudget = float64((Percent * float64(Budget)) / 100)
        item := StockQueryItem{StockName: Name, StockPercentage: Percent, StockBudget: _StockBudget}
        query.AddItem(item)
    }
    return query
}

/*
* Main Function
*/
func main() {
    Result = StockResults{[]StockResult{}}  //Initialize the Global Variable
    stocks := new(Stocks)
    server := rpc.NewServer()
    server.Register(stocks)
    server.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)
    listener, e := net.Listen("tcp", ":1234")
    if e != nil {
        log.Fatal("listen error:", e)
    }
    for {
        if conn, err := listener.Accept(); err != nil {
            log.Fatal("accept error: " + err.Error())
        } else {
            log.Printf("new connection established\n")
            go server.ServeCodec(jsonrpc.NewServerCodec(conn))  //Keep Serving the server
        }
    }
}
