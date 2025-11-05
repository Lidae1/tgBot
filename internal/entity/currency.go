package entity

type CurrencyName string

const (
	BTC CurrencyName = "BTC"
	ETH CurrencyName = "ETH"
)

var TokenList = map[int]CurrencyName{
	1: BTC,
	2: ETH,
}

type Price struct {
	Symbol  CurrencyName `json:"symbol"`
	Price   string       `json:"price"`
	Updated string       `json:"timestamp"`
}

type PriceResponse struct {
	BTC *Price `json:"BTC"`
	ETH *Price `json:"ETH"`
}
