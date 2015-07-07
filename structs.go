package okcoin

import (
	"encoding/json"
	"errors"
	//	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var channelsWithParams map[string]int

func init() {
	channelsWithParams = map[string]int{
		"ok_usd_realtrades":       1,
		"ok_spotusd_trade":        1,
		"ok_spotusd_cancel_order": 1,
		"ok_spotusd_userinfo":     1,
		"ok_spotusd_order_info":   1}
}

type Response struct {
	Channel string
	Data    interface{}
}

type Ticker struct {
	Channel string
	Data    TickerData
}

type TickerData struct {
	Buy       float64
	High      float64
	Last      float64
	Low       float64
	Sell      float64
	Timestamp uint64
	Vol       float64
}

type PriceAmount struct {
	Price  float64
	Amount float64
}

type Depth struct {
	Channel string
	Data    DepthData
}

type DepthData struct {
	Bids      []PriceAmount
	Asks      []PriceAmount
	Timestamp uint64
}

type Trades struct {
	Channel string
	Data    []Trade
}

type Trade struct {
	Id     uint64
	Price  float64
	Amount float64
	Time   string
	Bid    bool
}

func (r *Response) GetTicker() (t *Ticker, err error) {
	t = new(Ticker)
	t.Channel = r.Channel

	data, err := r.getDataAsMap()
	if err != nil {
		return t, err
	}
	err = convertMapToStruct(data, &t.Data)
	return t, err
}

func (r *Response) GetDepth() (d *Depth, err error) {
	d = new(Depth)
	d.Channel = r.Channel
	data, err := r.getDataAsMap()
	if err != nil {
		return d, err
	}
	err = convertMapToStruct(data, &d.Data)
	if err != nil {
		return d, err
	}
	d.Data.Bids, err = convertPriceAmounts(data["bids"])
	if err != nil {
		return
	}
	d.Data.Asks, err = convertPriceAmounts(data["asks"])
	if err != nil {
		return
	}
	return d, err
}

func (r *Response) GetTrades() (t *Trades, err error) {
	t = new(Trades)
	t.Channel = r.Channel

	var data [][]string

	buf, err := json.Marshal(r.Data)
	if err != nil {
		return t, err
	}
	err = json.Unmarshal(buf, &data)
	if err != nil {
		return t, err
	}
	t.Data = make([]Trade, len(data))
	for n, rec := range data {
		t.Data[n] = Trade{}
		if t.Data[n].Id, err = strconv.ParseUint(rec[0], 10, 64); err != nil {
			return t, err
		}
		if t.Data[n].Price, err = parseFloat64(rec[1]); err != nil {
			return t, err
		}
		if t.Data[n].Amount, err = parseFloat64(rec[2]); err != nil {
			return t, err
		}
		t.Data[n].Time = rec[3] //TODO translate to unix time
		t.Data[n].Bid = (rec[4] == "bid")
	}
	return t, err
}

func (r *Response) GetConverted() (interface{}, error) {

	if r.Channel == "ok_btcusd_ticker" || r.Channel == "ok_ltcusd_ticker" {
		return r.GetTicker()
	} else if r.Channel == "ok_btcusd_depth" || r.Channel == "ok_ltcusd_depth" {
		return r.GetDepth()
	} else if r.Channel == "ok_btcusd_trades_v1" {
		return r.GetTrades()
	}
	return nil, errors.New("Unrecognized response")
}

func (r *Response) getDataAsMap() (map[string]interface{}, error) {
	data, ok := r.Data.(map[string]interface{})
	if !ok {
		return nil, errors.New("Response data is not matched map[string]interface{} type")
	}
	return data, nil
}

func (r *Response) getDataAsArray() ([]interface{}, error) {
	data, ok := r.Data.([]interface{})
	if !ok {
		return nil, errors.New("Response data is not matched []interface{} type")
	}
	return data, nil
}

func getResponses(input []byte) ([]Response, error) {
	var rs []Response
	err := json.Unmarshal(input, &rs)
	if err != nil {
		return nil, err
	}
	return rs, nil
}

func convertPriceAmounts(data interface{}) ([]PriceAmount, error) {
	var ok bool
	var d1, d2 []interface{}
	err := errors.New("Error converting []PriceAmount")
	d1, ok = data.([]interface{})
	if !ok {
		return nil, err
	}
	ret := make([]PriceAmount, len(d1))
	for n, val := range d1 {
		d2, ok = val.([]interface{})
		if !ok {
			return nil, err
		}
		ret[n] = PriceAmount{}
		ret[n].Price, ok = d2[0].(float64)
		if !ok {
			return nil, err
		}
		ret[n].Amount, ok = d2[1].(float64)
		if !ok {
			return nil, err
		}
	}
	return ret, nil
}

func convertMapToStruct(data map[string]interface{}, result interface{}) error {

	resValue := reflect.ValueOf(result).Elem()

	for i := 0; i < resValue.NumField(); i++ {
		field := resValue.Field(i)
		fieldName := resValue.Type().Field(i).Name
		fieldTypeName := field.Type().Name()
		name := strings.ToLower(fieldName)

		switch fieldTypeName {
		case "float64":
			value, err := parseFloat64(data[name].(string))
			if err != nil {
				return err
			}
			field.SetFloat(value)
		case "uint64":
			value, err := strconv.ParseUint(data[name].(string), 10, 64)
			if err != nil {
				return err
			}
			field.SetUint(value)
		}
	}
	return nil
}

/*
func convertStringListToStruct(data []string, result interface{}) error {

	resValue := reflect.ValueOf(result).Elem()

	for i := 0; i < resValue.NumField(); i++ {
		field := resValue.Field(i)
		fieldTypeName := field.Type().Name()

		switch fieldTypeName {
		case "float64":
			value, err := parseFloat64(data[i])
			if err != nil {
				return err
			}
			field.SetFloat(value)
		case "uint64":
			value, err := strconv.ParseUint(data[i], 10, 64)
			if err != nil {
				return err
			}
			field.SetUint(value)
		case "string":
			field.SetString(data[i])
		}
	}
	return nil
}
*/

func parseFloat64(s string) (float64, error) {
	return strconv.ParseFloat(strings.Replace(s, ",", "", -1), 64)
}
