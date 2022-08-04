package main
import (
	stan "github.com/nats-io/stan.go"
  "fmt"
  "os"
)
var json1 = `{
  "order_uid": "123",
  "track_number": "WBILMTESTTRACK",
  "entry": "WBIL",
  "locale": "en",
  "internal_signature": "",
  "customer_id": "test",
  "delivery_service": "meest",
  "shardkey": "9",
  "sm_id": 99,
  "date_created": "2021-11-26T06:22:19Z",
  "oof_shard": "1",
  "delivery": {
    "name": "Test Testov",
    "phone": "+9720000000",
    "zip": "2639809",
    "city": "Kiryat Mozkin",
    "address": "Ploshad Mira 15",
    "region": "Kraiot",
    "email": "test@gmail.com"
  },
  "payment": {
    "transaction": "b563feb7b2b84b6test",
    "request_id": "",
    "currency": "USD",
    "provider": "wbpay",
    "amount": 1817,
    "payment_dt": 1637907727,
    "bank": "alpha",
    "delivery_cost": 1500,
    "goods_total": 317,
    "custom_fee": 0
  },
  "items": [
    {
      "chrt_id": 9934930,
      "track_number": "WBILMTESTTRACK",
      "price": 453,
      "rid": "ab4219087a764ae0btest",
      "name": "Mascaras",
      "sale": 30,
      "size": "0",
      "total_price": 317,
      "nm_id": 2389212,
      "brand": "Vivienne Sabo",
      "status": 202
    }
  ]
}`
var json2 = `{
  "order_uid": "uidtest123",
  "track_number": "WBILMTESTTRACK2",
  "entry": "WBIL",
  "locale": "en",
  "internal_signature": "",
  "customer_id": "test",
  "delivery_service": "meest",
  "shardkey": "9",
  "sm_id": 10,
  "date_created": "2021-09-26T06:22:19Z",
  "oof_shard": "1",
  "delivery": {
    "name": "Test Testov",
    "phone": "+97233300000",
    "zip": "2639809",
    "city": "Kiryat Mozkin",
    "address": "Ploshad Mira 15",
    "region": "Kraiot",
    "email": "test@gmail.com"
  },
  "payment": {
    "transaction": "b5ssseb7b2b84b6test",
    "request_id": "",
    "currency": "USD",
    "provider": "wbpay",
    "amount": 1817,
    "payment_dt": 1633307727,
    "bank": "alpha",
    "delivery_cost": 1500,
    "goods_total": 317,
    "custom_fee": 0
  },
  "items": [
    {
      "chrt_id": 9334930,
      "track_number": "WBILMTESTTRACK2",
      "price": 453,
      "rid": "ab38t5719087a764ae0btest",
      "name": "Mascaras",
      "sale": 30,
      "size": "0",
      "total_price": 317,
      "nm_id": 2389212,
      "brand": "Vivienne Sabo",
      "status": 202
    },
    {
      "chrt_id": 93222230,
      "track_number": "WBILMTESTTRACK2",
      "price": 453,
      "rid": "1111t5719087a764ae0btest",
      "name": "Mascaras",
      "sale": 30,
      "size": "0",
      "total_price": 317,
      "nm_id": 23800012,
      "brand": "Vivienne Sabo",
      "status": 202
    },
    {
      "chrt_id": 1111930,
      "track_number": "WBILMTESTTRACK2",
      "price": 453,
      "rid": "11111t5719087a764ae0btest",
      "name": "Mascaras",
      "sale": 30,
      "size": "0",
      "total_price": 317,
      "nm_id": 1111212,
      "brand": "Vivienne Sabo",
      "status": 202
    }
  ]
}`
func main() {
	stanConn, err := stan.Connect("test-cluster", "simple")
  if err != nil{
    fmt.Println("Ошибка при подключении к NATS-Streaming!")
    os.Exit(1)
  }
  defer stanConn.Close()
  fmt.Println("NATS-Streaming: начало рассылки")
	stanConn.Publish("service", []byte(json1))
  stanConn.Publish("service", []byte(json2))
  stanConn.Publish("service", []byte("XXXXYYYYNNNNN"))
  fmt.Println("Рассылка завершена!")
}
