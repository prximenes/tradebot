package main

import (
	"crypto/hmac"
	"crypto/sha512"
	"fmt"
	"os"
	"reflect"

	//"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/yanzay/tbot"
)

//
const baseURL string = "https://bleutrade.com/api/v2/"

var Key, faSecret, roSecret string //Variaveis globais que guardam chaves de autenticacao da API
var quantity, comments, market, address, currency, rate string

func computeHmac512(message string, secret string) string { //Computa chave de autenticacao usando hmac sha512
	enKey := []byte(secret)
	h := hmac.New(sha512.New, enKey)
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}
func digestBody(resp *http.Response) map[string]interface{} { //Trata corpo de uma mensagem de resposta HTTP e retorna a resposta como JSON
	body, readErr := ioutil.ReadAll(resp.Body) //Le corpo da mensagem
	if readErr != nil {
		log.Fatal(readErr)
	}
	var ret map[string]interface{} //Cria nova interface que sera usada como "struct" para JSON

	jsonErr := json.Unmarshal(body, &ret) //Parse da informacao lida
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}
	return ret
}

func getResults(resp *http.Response) []map[string]interface{} { //Trata corpo de uma mensagem de resposta HTTP e retorna a resposta como JSON
	body, readErr := ioutil.ReadAll(resp.Body) //Le corpo da mensagem
	if readErr != nil {
		log.Fatal(readErr)
	}
	var jBody map[string]interface{} //Cria nova interface que sera usada como "struct" para JSON
	//fmt.Println(string(body))
	jsonErr := json.Unmarshal(body, &jBody) //Parse da informacao lida
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}
	//if jBody["sucess"] == true {
	if jBody["result"] != nil { //Se resultados foram retornados
		s := reflect.ValueOf(jBody["result"]) //Usa reflect para recuperar resultados como slice de maps
		if s.Kind() != reflect.Slice {
			panic("InterfaceSlice() given a non-slice type")
		}

		results := make([]map[string]interface{}, s.Len())
		for i := 0; i < s.Len(); i++ {
			results[i] = s.Index(i).Interface().(map[string]interface{}) //Type asserting para transformar Interface em Map[string]
		}

		return results
	}
	//}
	return nil
}

func GetMarkets() []map[string]interface{} {
	resp, err := http.Get(baseURL + "public/getmarkets")
	if err != nil {
		log.Fatal(err)
	}
	return getResults(resp)
}

func GetBalance(currency string) []map[string]interface{} { //Retorna interface com informacões de saldo da carteira
	var uri string
	uri = baseURL + "account/getbalances?currencies=" + currency + "&apikey=" + Key //Constroi URI

	var sign = computeHmac512(uri, roSecret) //Computa chave de autenticacao

	resp, err := http.Get(uri + "&apisign=" + sign) //Adiciona chave na URI e faz requisicao
	if err != nil {
		log.Fatal(err)
	}

	return getResults(resp)
}

func GetBalances(currency []string) []map[string]interface{} {
	var uri string //Construcao da URI
	currencies := strings.Join(currency, ";")
	if len(currencies) == 0 {
		uri = baseURL + "account/getbalances?" + "apikey=" + Key
	} else {
		uri = baseURL + "account/getbalances?currencies=" + currencies + "&apikey=" + Key
	}

	var sign = computeHmac512(uri, roSecret)

	resp, err := http.Get(uri + "&apisign=" + sign)
	if err != nil {
		log.Fatal(err)
	}
	return getResults(resp)
}

func GetOrders(market, orderstatus, ordertype, depth string) []map[string]interface{} {
	var uri = baseURL + "account/getorders?"
	var check bool
	if len(market) > 0 {
		check = true
		uri += "market=" + market
	}
	if len(orderstatus) > 0 {
		if check {
			uri += "&"
		}
		check = true
		uri += "orderstatus=" + orderstatus
	}
	if len(ordertype) > 0 {
		if check {
			uri += "&"
		}
		check = true
		uri += "ordertype=" + ordertype
	}
	if len(depth) > 0 {
		if check {
			uri += "&"
		}
		check = true
		uri += "depth=" + depth
	}

	uri += "&apikey=" + Key

	var sign = computeHmac512(uri, roSecret)
	resp, err := http.Get(uri + "&apisign=" + sign)
	if err != nil {
		log.Fatal(err)
	}

	return getResults(resp)
}

func Withdraw(currency, quantity, address, comments string) bool {
	var uri = baseURL + "account/withdraw?" + "currency=" + currency + "&quantity=" + quantity + "&address=" + address
	if comments != "" {
		uri += "&comments=" + comments
	}
	uri += "&apikey=" + Key

	var sign = computeHmac512(uri, faSecret)

	resp, err := http.Get(uri + "&apisign=" + sign)
	if err != nil {
		log.Fatal(err)
	}

	if (digestBody(resp)["success"].(string)) == "true" {
		return true
	}
	return false
}

func BuyCoin(market, rate, quantity, comments string) []map[string]interface{} {
	var uri = baseURL + "account/buylimit?" + "market=" + market + "&rate=" + rate + "&quantity=" + quantity
	if comments != "" {
		uri += "&comments=" + comments
	}
	uri += "&apikey=" + Key

	var sign = computeHmac512(uri, faSecret)

	resp, err := http.Get(uri + "&apisign=" + sign)
	if err != nil {
		log.Fatal(err)
	}

	return getResults(resp)
}
func SellCoin(market, rate, quantity, comments string) []map[string]interface{} {
	var uri = baseURL + "account/selllimit?" + "market=" + market + "&rate=" + rate + "&quantity=" + quantity
	if comments != "" {
		uri += "&comments=" + comments
	}
	uri += "&apikey=" + Key

	var sign = computeHmac512(uri, faSecret)

	resp, err := http.Get(uri + "&apisign=" + sign)
	if err != nil {
		log.Fatal(err)
	}

	return getResults(resp)
}

/*func init() {
	txt, readErr := ioutil.ReadFile("k.txt")
	if readErr != nil {
		log.Fatal(readErr)
	}
	var access map[string]interface{}

	jsonErr := json.Unmarshal(txt, &access)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	Key = access["Key"].(string)
	faSecret = access["FASecret"].(string)
	roSecret = access["ROSecret"].(string)
}*/

//bot_func
func BalanceHandler(message *tbot.Message) {
	currencies := strings.Split(strings.ToUpper(message.Vars["currency"]), ";")
	fmt.Println(currencies)
	balances := GetBalances(currencies)
	var reply string
	var balance map[string]interface{}
	fmt.Println(balances)
	for i := range balances {
		balance = balances[i]
		//for field := range balance {
		reply += "Saldo" + ": " + balance["Available"].(string) + "; "
		//}
		reply += "\n"
	}
	message.Replyf(reply)
}

//bot func
func loginHandler(message *tbot.Message) {
	Key = message.Vars["key"]
	roSecret = message.Vars["ro_key"]
	faSecret = message.Vars["fa_key"]

	message.Replyf("Logado")
}

/*func KeyHandler(message *tbot.Message) {
	// Message contain it's varialbes from curly brackets
	Key = message.Vars["my_key"]
	fmt.Println(Key)
	message.Reply(message.Vars["my_key"])
}

func FASecretHandler(message *tbot.Message) {
	// Message contain it's varialbes from curly brackets
	faSecret = message.Vars["my_fa"]
	message.Reply(message.Vars["my_fa"])
}

func ROSecretHandler(message *tbot.Message) {
	// Message contain it's varialbes from curly brackets
	roSecret = message.Vars["my_ro"]
	message.Reply(message.Vars["my_ro"])
}*/

func SellHandler(message *tbot.Message) {
	market = message.Vars["market"]
	rate = message.Vars["rate"]
	quantity = message.Vars["quantity"]
	comments = message.Vars["comments"]
	SellCoin(market, rate, quantity, comments)
	message.Replyf("Operação Realizada com sucesso")

}

func BuyHandler(message *tbot.Message) {
	market = message.Vars["market"]
	rate = message.Vars["rate"]
	quantity = message.Vars["quantity"]
	comments = message.Vars["comments"]
	BuyCoin(market, rate, quantity, comments)
}

func KeyboardHandler(message *tbot.Message) {
	buttons := [][]string{
		{"/help"},
		//{"Another", "Row"},
	}
	message.ReplyKeyboard("Help!", buttons)
}

func TransfHandler(message *tbot.Message) {
	currency = message.Vars["currency"]
	address = message.Vars["address"]
	quantity = message.Vars["quantity"]
	comments = message.Vars["comments"]
	//fmt.Println(currency)
	//fmt.Println(address)
	//fmt.Println(quantity)
	//fmt.Println(comments)
	err := Withdraw(currency, quantity, address, comments)
	fmt.Println(err)
	if err == true {
		message.Replyf("Operação Realizada com sucesso")
	} else {
		message.Replyf("Operação não realizada")
	}
}

func main() {
	//Telegram_bot config
	bot, err := tbot.NewServer(os.Getenv("TELEGRAM_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	bot.HandleFunc("/login {key}, {ro_key}, {fa_key}", loginHandler)
	bot.HandleFunc("/saldo {currency}", BalanceHandler)
	bot.HandleFunc("/sell {market}, {rate}, {quantity}, {comments}", SellHandler)
	bot.HandleFunc("/buy {market}, {rate}, {quantity}, {comments}", BuyHandler)
	bot.HandleFunc("/transf {currency}, {quantity}, {address}, {comments}", TransfHandler)
	bot.HandleFunc("/keyboard", KeyboardHandler)

	bot.ListenAndServe()

	//balances := getBalances([]string{"BTC", "DOGE"})
	//fmt.Println("\n \t ", balances)
	/*balance := getBalance("")
	fmt.Println("\n \t ", balance)
	balance = getBalance("BTC")
	fmt.Println("\n \t ", balance)*/
	//orders := getOrders("ALL", "All", "All", "")
	//fmt.Println(orders)

	//transf := withdraw("DOGE", "11", "DSjMqGCZg6VtpQQqnbLkDD1gHf4X42WLKs", "jeri")
	//fmt.Println(transf)
	//sell := sellCoin()
	//buy := buyCoin("DOGE_BTC")
}
