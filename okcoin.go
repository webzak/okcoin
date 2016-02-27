//Package okcoin implements okcoin websocket api (https://www.okcoin.com/about/ws_getStarted.do)
package okcoin

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/gorilla/websocket"
)

//Addresses for USD and CNY OKCoin markets
const (
	USDWsAPIURL = "wss://real.okcoin.com:10440/websocket/okcoinapi"
	CNYWsAPIURL = "wss://real.okcoin.cn:10440/websocket/okcoinapi"
)

//WsAPI represents websocket api
type WsAPI struct {
	ws     *websocket.Conn
	pubKey string
	prvKey string
}

//Req contains the request data
type Req struct {
	//Channel specifies the channel name
	Channel string
	//On true - enable channel, false disable channel
	On bool
	//Params - parameters specific for particular channel
	Params map[string]string
}

//NewReq creates new request *Req
func NewReq(channel string, on bool, params ...map[string]string) *Req {
	if params != nil {
		return &Req{channel, on, params[0]}
	}
	if _, ok := channelsWithParams[channel]; ok {
		return &Req{channel, on, make(map[string]string)}
	}
	return &Req{channel, on, nil}
}

//NewWsAPI creates new *WsAPI providing keys pair
func NewWsAPI(publicKey, privateKey string) (*WsAPI, error) {
	if len(publicKey) == 0 || len(privateKey) == 0 {
		return nil, errors.New("Keys are not valid")
	}
	ws := new(WsAPI)
	ws.pubKey = publicKey
	ws.prvKey = privateKey
	return ws, nil
}

//Connect establishes websocket connection
func (w *WsAPI) Connect(symbol string) (err error) {
	dialer := websocket.Dialer{ReadBufferSize: 10000, WriteBufferSize: 1000}
	if symbol == "btc_cny" {
		w.ws, _, err = dialer.Dial(CNYWsAPIURL, nil)
	} else if symbol == "btc_usd" {
		w.ws, _, err = dialer.Dial(USDWsAPIURL, nil)
	}
	return err
}

//Close closes websocket connection
func (w *WsAPI) Close() error {
	return w.ws.Close()
}

//Ping sends keep alive message and verifies server response
//if error returned reconnect is required to continue operation
func (w *WsAPI) Ping() error {
	err := w.ws.WriteMessage(websocket.TextMessage, []byte(`{"event":"ping"}`))
	if err != nil {
		return err
	}
	_, ret, err := w.ws.ReadMessage()
	if err != nil {
		return err
	}
	if string(ret) != `{"event":"pong"}` {
		err = errors.New("Ping error. Response: " + string(ret))
	}
	return err
}

//Send sends request *Req to server
//It is possible to send several requests simultaneously
func (w *WsAPI) Send(reqs ...*Req) error {
	if reqs == nil {
		return errors.New("Empty requests")
	}
	message, err := w.createMessage(reqs)
	if err != nil {
		return err
	}
	return w.ws.WriteMessage(websocket.TextMessage, []byte(message))
}

//Read returns raw response from api
func (w *WsAPI) Read() ([]byte, error) {
	_, data, err := w.ws.ReadMessage()
	if err != nil {
		return nil, err
	}
	return data, nil
}

//ReadResponses returns partically parsed api responses
//The Response.Channel value can be used to detect the response contents
//The method is useful for cases when some responses might be ignored
func (w *WsAPI) ReadResponses() ([]Response, error) {
	data, err := w.Read()
	if err != nil {
		return nil, err
	}
	return getResponses(data)
}

//ReadConverted returns the response converted to appropriate structure
//The possible return value type is one of: *Ticker, *Depth, *Trades
func (w *WsAPI) ReadConverted() ([]interface{}, error) {
	rs, err := w.ReadResponses()
	if err != nil {
		return nil, err
	}
	ret := make([]interface{}, len(rs))
	for n, r := range rs {
		val, err := r.GetConverted()
		if err != nil {
			return nil, err
		}
		ret[n] = val
	}
	return ret, nil
}

//create message by combining request events
func (w *WsAPI) createMessage(reqs []*Req) (string, error) {
	chunks := make([]string, len(reqs))
	for n, req := range reqs {
		m, err := w.createEvent(req)
		if err != nil {
			return "", err
		}
		chunks[n] = m
	}
	var ret string
	if len(chunks) == 1 {
		ret = chunks[0]
	} else {
		ret = "[" + strings.Join(chunks, ",") + "]"
	}
	return ret, nil
}

//create single event from request
func (w *WsAPI) createEvent(r *Req) (string, error) {
	var op, message string
	if r.On {
		op = "add"
	} else {
		op = "remove"
	}
	if r.Params == nil {
		message = fmt.Sprintf(`{"event":"%sChannel","channel":"%s"}`, op, r.Channel)
	} else {
		params, err := w.prepareParams(r.Params)
		if err != nil {
			return "", err
		}
		paramsJSON, err := json.Marshal(params)
		if err != nil {
			return "", err
		}
		message = fmt.Sprintf(`{"event":"%sChannel","channel":"%s", "parameters": %s}`, op, r.Channel, string(paramsJSON))
	}
	return message, nil
}

// adds api key and private key sign for parameters
func (w *WsAPI) prepareParams(params map[string]string) (map[string]string, error) {
	params["api_key"] = w.pubKey
	ps := prepareParamString(params)
	md5 := signParamString(ps, w.prvKey)
	params["sign"] = md5
	return params, nil
}

// sorts alfabetically and creates the string for sign
func prepareParamString(params map[string]string) string {
	pairs := make([]string, len(params))
	i := 0
	for key, value := range params {
		pairs[i] = key + "=" + value
		i++
	}
	sort.Strings(pairs)
	return strings.Join(pairs, "&")
}

// adds private key and calculates md5
func signParamString(s, prvKey string) string {
	signedStr := s + "&secret_key=" + prvKey
	return strings.ToUpper(fmt.Sprintf("%x", md5.Sum([]byte(signedStr))))
}
