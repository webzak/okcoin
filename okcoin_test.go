package okcoin

import (
	"errors"
	//"fmt"
	"github.com/stretchr/testify/assert"
	ini "github.com/vaughan0/go-ini"
	"testing"
)

func TestInit(t *testing.T) {

	pub := "aaa"
	prv := "bbb"
	api, err := NewWsApi(pub, prv)
	assert := assert.New(t)
	assert.Nil(err)

	assert.Equal(api.pubKey, pub)

	assert.Equal(api.prvKey, prv)

	api, err = NewWsApi(pub, "")
	assert.NotNil(err)
}

/*
func TestChannel(t *testing.T) {
	pub, prv, err := loadKeys()
	assert.Nil(t, err)

	api, err := NewWsApi(pub, prv)
	err = api.Connect()
	assert.Nil(t, err)
	defer api.Close()

	//channels := []string{"ok_btcusd_ticker", "ok_btcusd_depth"}
	//err = api.Send(NewReq("ok_btcusd_ticker", true), NewReq("ok_btcusd_trades_v1", true))
	assert.Nil(t, err)
	err = api.Send(&Req{"ok_btcusd_ticker", true, nil})
	//err = api.AddChannel("ok_btcusd_trades_v1")"ok_btcusd_ticker", "ok_btcusd_depth"
	assert.Nil(t, err)
	i := 1
	for i < 2 {
		ret, err := api.Read()
		assert.Nil(t, err)
		fmt.Println(string(ret))
		i++

	}

	err = api.Send(NewReq("ok_btcusd_deps", false))
	assert.Nil(t, err)


	//	err = api.Send(NewReq("ok_spotusd_userinfo", true))
	//	assert.Nil(t, err)
	//	ret, err := api.Read()
	//	assert.Nil(t, err)
	//	fmt.Println(string(ret))

}
*/

func Test_createEvent(t *testing.T) {
	api, _ := NewWsApi("pub", "priv")
	req := &Req{"ok_btcusd_ticker", true, nil}
	m, err := api.createEvent(req)
	assert.Nil(t, err)
	assert.Equal(t, `{"event":"addChannel","channel":"ok_btcusd_ticker"}`, m)

	req = &Req{"ok_spotusd_userinfo", true, make(map[string]string)}
	m, err = api.createEvent(req)
	assert.Nil(t, err)
	assert.Equal(t, `{"event":"addChannel","channel":"ok_spotusd_userinfo", "parameters": {"api_key":"pub","sign":"75460F59332AE8EBA0012043A7CF9132"}}`, m)

	req = NewReq("ok_spotusd_userinfo", true, map[string]string{"x": "y"})
	m, err = api.createEvent(req)
	assert.Nil(t, err)
	assert.Equal(t, `{"event":"addChannel","channel":"ok_spotusd_userinfo", "parameters": {"api_key":"pub","sign":"98C2C8EDF767B6EF3CAA47290C59D747","x":"y"}}`, m)
}

func Test_createMessage(t *testing.T) {
	api, _ := NewWsApi("pub", "priv")
	req := NewReq("ok_btcusd_ticker", true)
	m, err := api.createMessage([]*Req{req})
	assert.Nil(t, err)
	assert.Equal(t, `{"event":"addChannel","channel":"ok_btcusd_ticker"}`, m)
	req2 := NewReq("ok_spotusd_userinfo", true)
	m, err = api.createMessage([]*Req{req, req2})
	assert.Nil(t, err)
	assert.Equal(t, `[{"event":"addChannel","channel":"ok_btcusd_ticker"},{"event":"addChannel","channel":"ok_spotusd_userinfo", "parameters": {"api_key":"pub","sign":"75460F59332AE8EBA0012043A7CF9132"}}]`, m)
}

func Test_signParamString(t *testing.T) {
	s := "amount=0.02&api_key=c821db84-6fbd-11e4-a9e3-c86000d26d7c&price=50&symbol=btc_usd&type=buy"
	ret := signParamString(s, "123")
	assert.Equal(t, "663BC5D28401A84693BEBFD2A0386645", ret)
}

func Test_prepareParamString(t *testing.T) {

	m := make(map[string]string)
	m["one"] = "1"
	m["eight"] = "8"
	m["five"] = "5"
	ret := prepareParamString(m)
	assert.Equal(t, "eight=8&five=5&one=1", ret)
}

func Test_prepareParams(t *testing.T) {
	pubKey := "c821db84-6fbd-11e4-a9e3-c86000d26d7c"
	prvKey := "123"
	api, _ := NewWsApi(pubKey, prvKey)
	p := make(map[string]string)
	p["symbol"] = "btc_usd"
	p["type"] = "buy"
	p["amount"] = "0.02"
	p["price"] = "50"

	assert := assert.New(t)
	ret, err := api.prepareParams(p)
	assert.Nil(err)
	assert.Equal("663BC5D28401A84693BEBFD2A0386645", ret["sign"])
	assert.Equal(pubKey, ret["api_key"])
	//fmt.Println(ret)
}

func loadKeys() (pub, prv string, err error) {
	var file ini.File
	file, err = ini.LoadFile("keys.ini")
	if err != nil {
		return
	}
	var ok bool
	pub, ok = file.Get("keys", "public")
	if !ok {
		err = errors.New("pubkey missed")
		return
	}
	prv, ok = file.Get("keys", "private")
	if !ok {
		err = errors.New("privkey missed")
		return
	}
	return
}

func TestTicker(t *testing.T) {
	assert := assert.New(t)
	in := []byte(`[{"channel":"ok_btcusd_ticker","data":{"buy":"253.36","high":"256.52","last":"253.36","low":"250.92","sell":"253.4","timestamp":"1435925683940","vol":"8,705.22"}}]`)
	ret, err := getResponses(in)
	assert.Nil(err)
	//fmt.Printf("%#v", ret)
	ticker, err := ret[0].GetTicker()
	assert.Nil(err)
	//fmt.Printf("%#v", ticker)
	assert.Equal(253.36, ticker.Data.Buy)
	assert.Equal(8705.22, ticker.Data.Vol)
	assert.Equal(uint64(1435925683940), ticker.Data.Timestamp)

	val, err := ret[0].GetConverted()
	assert.Nil(err)
	//fmt.Printf("%#v", val)
	switch nval := val.(type) {
	case *Ticker:
		assert.Equal(253.36, nval.Data.Buy)
		assert.Equal(8705.22, nval.Data.Vol)
		assert.Equal(uint64(1435925683940), nval.Data.Timestamp)
	}
}

func TestDepth(t *testing.T) {
	assert := assert.New(t)
	in := []byte(`[{"channel":"ok_btcusd_depth","data":{"bids":[[255.5,3.066],[255.49,0.175],[255.42,0.709],[255.41,0.125],[255.39,7.316],[255.34,0.125],[255.33,0.125],[255.31,0.115],[255.29,0.125],[255.27,0.125],[255.25,0.109],[255.24,1.54],[255.2,0.36],[255.18,0.01],[255.16,1.458],[255.11,1.839],[255.1,6.552],[254.98,1.536],[254.97,1630],[254.96,36.264]],"asks":[[256.18,0.01],[256,2.099],[255.98,4.01],[255.84,10],[255.82,0.36],[255.78,0.01],[255.77,3.883],[255.76,0.695],[255.75,0.238],[255.74,0.209],[255.73,0.278],[255.72,0.371],[255.71,0.222],[255.7,0.232],[255.69,0.239],[255.68,0.215],[255.67,0.295],[255.63,0.05],[255.61,0.05],[255.51,0.5]],"timestamp":"1436012091204"}}]`)
	ret, err := getResponses(in)
	assert.Nil(err)
	//fmt.Printf("%#v\n", ret)

	depth, err := ret[0].GetDepth()
	assert.Nil(err)
	//fmt.Printf("%#v\n", depth)

	assert.Equal(255.5, depth.Data.Bids[0].Price)
	assert.Equal(2.099, depth.Data.Asks[1].Amount)

	val, err := ret[0].GetConverted()
	assert.Nil(err)
	//fmt.Printf("%#v", val)

	switch nval := val.(type) {
	case *Depth:
		assert.Equal(255.5, nval.Data.Bids[0].Price)
		assert.Equal(2.099, nval.Data.Asks[1].Amount)
	}
}

func TestTrades(t *testing.T) {
	assert := assert.New(t)
	in := []byte(`[{"channel":"ok_btcusd_trades_v1","data":[["24307225","257.46","0.05","22:32:24","ask"],["24307282","257.46","0.17","22:32:44","ask"],["24307400","257.46","0.78","22:33:10","ask"],["24307513","257.46","0.52","22:33:34","ask"],["24307634","257.49","10.15","22:33:59","ask"],["24307727","257.5","0.3","22:34:22","ask"],["24307729","257.46","0.83","22:34:22","ask"],["24307868","257.46","0.62","22:34:52","ask"],["24307980","257.46","0.16","22:35:21","ask"],["24308067","257.46","0.65","22:35:43","ask"],["24308188","257.46","0.92","22:36:13","ask"],["24308293","257.46","0.05","22:36:37","ask"],["24308316","257.53","0.729","22:36:39","bid"],["24308318","257.54","0.148","22:36:39","bid"],["24308320","257.57","0.148","22:36:39","bid"],["24308322","257.59","0.148","22:36:39","bid"],["24308324","257.61","0.148","22:36:39","bid"],["24308326","257.63","0.148","22:36:39","bid"],["24308327","257.64","0.705","22:36:39","ask"],["24308395","257.65","0.169","22:36:55","ask"],["24308397","257.64","0.711","22:36:55","ask"],["24308516","257.65","0.148","22:37:22","ask"]]}]`)
	ret, err := getResponses(in)
	assert.Nil(err)
	//fmt.Printf("%#v\n", ret)
	trades, err := ret[0].GetTrades()
	assert.Nil(err)

	//fmt.Printf("%#v\n", trades)

	assert.Equal(uint64(24307225), trades.Data[0].Id)
	assert.Equal(257.46, trades.Data[1].Price)
	assert.Equal(0.78, trades.Data[2].Amount)
	assert.Equal("22:33:10", trades.Data[2].Time)
	assert.Equal(false, trades.Data[2].Bid)

	val, err := ret[0].GetConverted()
	assert.Nil(err)

	//fmt.Printf("%#v", val)
	trades, ok := val.(*Trades)
	assert.True(ok)
	//fmt.Printf("%#v", trades)
	assert.Equal(uint64(24307225), trades.Data[0].Id)
	assert.Equal(257.46, trades.Data[1].Price)
	assert.Equal(0.78, trades.Data[2].Amount)
	assert.Equal("22:33:10", trades.Data[2].Time)
	assert.Equal(false, trades.Data[2].Bid)
}

/*
func TestRead(t *testing.T) {
	pub, prv, err := loadKeys()

	api, err := NewWsApi(pub, prv)
	err = api.Connect()
	defer api.Close()

	err = api.Ping()
	assert.Nil(t, err)
	err = api.Send(&Req{"ok_btcusd_trades_v1", true, nil})
	assert.Nil(t, err)
	i := 1
	for i < 5 {
		ret, err := api.Read()
		assert.Nil(t, err)
		fmt.Printf("%s\n", ret)
		i++
	}
}
*/
/*
func TestReadResponses(t *testing.T) {
	pub, prv, err := loadKeys()

	api, err := NewWsApi(pub, prv)
	err = api.Connect()
	defer api.Close()

	err = api.Ping()
	assert.Nil(t, err)
	err = api.Send(&Req{"ok_btcusd_ticker", true, nil})
	assert.Nil(t, err)
	i := 1
	for i < 5 {
		ret, err := api.ReadResponses()
		assert.Nil(t, err)
		fmt.Printf("%#v\n", ret)
		i++
	}
}
*/
/*
func TestReadConverted(t *testing.T) {
	pub, prv, err := loadKeys()

	api, err := NewWsApi(pub, prv)
	err = api.Connect()
	defer api.Close()

	err = api.Ping()
	assert.Nil(t, err)
	err = api.Send(&Req{"ok_ltcusd_ticker", true, nil})
	assert.Nil(t, err)
	i := 1
	for i < 5 {
		ret, err := api.ReadConverted()
		assert.Nil(t, err)
		for _, data := range ret {
			if ticker, ok := data.(*Ticker); ok {
				fmt.Printf("%#v\n", ticker)
			}
		}
		i++
	}
}
*/
