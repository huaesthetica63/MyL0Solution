package main
import (
	"fmt"
	"os"
	"github.com/joho/godotenv"
	"context"
	"encoding/json"
	"net/http"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/patrickmn/go-cache"
	"github.com/nats-io/stan.go"
	"github.com/gin-gonic/gin"
)
//структуры для парсинга json, которые были сделаны по заданной модели
type deliveryInfo struct {
	Name string `json:"name"`
	Phone string `json:"phone"`
	Zip string `json:"zip"`
	City string `json:"city"`
	Address string `json:"address"`
	Region string `json:"region"`
	Email string `json:"email"`
}
type paymentInfo struct {
	Transaction string `json:"transaction"`
	Request_id string `json:"request_id"`
	Currency string `json:"currency"`
	Provider string `json:"provider"`
	Amount int `json:"amount"`
	Payment_dt int64 `json:"payment_dt"`
	Bank string `json:"bank"`
	Delivery_cost int `json:"delivery_cost"`
	Goods_total int `json:"goods_total"`
	Custom_fee int `json:"custom_fee"`
}
type itemInfo struct {
	Chrt_id int64 `json:"chrt_id"`
	Track_number string `json:"track_number"`
	Price int `json:"price"`
	Rid string `json:"rid"`
	Name string `json:"name"`
	Sale int `json:"sale"`
	Size string `json:"size"`
	Total_price int `json:"total_price"`
	Nm_id int64 `json:"nm_id"`
	Brand string `json:"brand"`
	Status int `json:"status"`
}
type orderInfo struct {
	Order_uid string `json:"order_uid"`
	Track_number string `json:"track_number"`
	Entry string `json:"entry"`
	Delivery deliveryInfo `json:"delivery"`
	Payment paymentInfo `json:"payment"`
	Items []itemInfo `json:"items"`
	Locale string `json:"locale"`
	Internal_signature string `json:"internal_signature"`
	Customer_id string `json:"customer_id"`
	Delivery_service string `json:"delivery_service"`
	Shardkey string `json:"shardkey"`
	Sm_id int64 `json:"sm_id"`
	Date_created string `json:"date_created"`
	Oof_shard string `json:"oof_shard"`
}
//envStruct хранит все переменные окружения, полученные с помощью godotenv
type envStruct struct {
	db_name string	//имя базы данных
	host_name string	//хост для размещения бд
	username string	//имя пользователя в бд
	password string	//его пароль
	postgres_port string	//порт в postgres по умолчанию 5432 или 5433
	http_port string //8080 для http-сервера
}
//инициализация переменных окружения из env файла
func initEnvironment() (envStruct, error){
	var envRes envStruct
	if err := godotenv.Load(); err != nil {
		fmt.Println("Ошибка инициализации окружения!")
		return envRes, err
	}
	//теперь можно пользоваться os.getenv для получения переменных
	envRes.db_name = os.Getenv("DB_NAME")
	envRes.host_name = os.Getenv("HOST_NAME")
	envRes.username = os.Getenv("USERNAME")
	envRes.password=os.Getenv("PASSWORD")
	envRes.postgres_port=os.Getenv("POSTGRES_PORT")
	envRes.http_port=os.Getenv("HTTP_PORT")
	fmt.Println("Переменные окружения загружены успешно!")
	return envRes,nil
}
//ФУНКЦИИ ДЛЯ ЗАПИСИ ДАННЫХ В POSTGRES ИЗ РАССЫЛКИ NATS
// записываем данные в таблицу orderInfo
func insertOrderInfo(conn *pgx.Conn, data orderInfo) (err error) {
	sqlQueryStr := `INSERT INTO orderInfo (order_uid, track_number, entry, locale, internal_signature,customer_id, delivery_service, shardkey, sm_id, date_created,oof_shard)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) returning order_uid`
	var order_uid string
	resOp := conn.QueryRow(context.Background(), sqlQueryStr, data.Order_uid, data.Track_number, data.Entry,
	data.Locale,data.Internal_signature,data.Customer_id, data.Delivery_service, data.Shardkey, data.Sm_id, data.Date_created, data.Oof_shard,
	).Scan(&order_uid);
	if resOp != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			return pgErr
		}
	}
	return nil
}
// Записываем данные в таблицу deliveryInfo
func insertDeliveryInfo(conn *pgx.Conn, data orderInfo) (err error) {
	sqlQueryStr := `INSERT INTO deliveryInfo (name, phone, zip, city, address, region, email, order_id)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8) returning delivery_id`
	del := data.Delivery //для удобства создаем отдельную переменную
	var delivery_id int
	resOp:=conn.QueryRow(context.Background(), sqlQueryStr,del.Name,del.Phone,del.Zip,del.City,del.Address,
	del.Region,del.Email,data.Order_uid).Scan(&delivery_id);
	if resOp!=nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			fmt.Println(pgErr)
			return pgErr
		}
	}
	return nil
}
// Записываем данные в таблицу paymentInfo
func insertPaymentInfo(conn *pgx.Conn, data orderInfo) ( err error) {
	sqlQueryStr := `INSERT INTO paymentInfo (transaction, request_id, currency, provider,
    amount, payment_dt, bank, delivery_cost, goods_total, custom_fee,order_id)
	VALUES ($1,$2,$3,$4,$5,$6,$7, $8, $9, $10, $11) returning payment_id`
	pay := data.Payment //аналогично используем не основной объект, а отдельно payment
	var payment_id int
	resOp := conn.QueryRow(context.Background(), sqlQueryStr, pay.Transaction, pay.Request_id, pay.Currency,
	pay.Provider, pay.Amount, pay.Payment_dt, pay.Bank, pay.Delivery_cost, pay.Goods_total,
	pay.Custom_fee, data.Order_uid).Scan(&payment_id);
	if resOp != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			fmt.Println(pgErr)
			return pgErr
		}
	}
	return nil
}
// Вставка слайса items в таблицу itemInfo
func insertItemInfo(conn *pgx.Conn, data orderInfo) ( err error) {
	sqlQueryStr := `INSERT INTO itemInfo(chrt_id,track_number,price,rid,name,sale,
	size, total_price,nm_id,brand,status,order_id)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) returning item_id`
	itemsCount := len(data.Items)
	for i := 0; i < itemsCount; i++ { //вставляем аналогично предыдущим случаям, но для всех items
		tempItem := data.Items[i]
		var item_id int
		resOp := conn.QueryRow(context.Background(), sqlQueryStr, tempItem.Chrt_id,
		tempItem.Track_number, tempItem.Price, tempItem.Rid, tempItem.Name,
		tempItem.Sale, tempItem.Size, tempItem.Total_price, tempItem.Nm_id,
		tempItem.Brand, tempItem.Status, data.Order_uid).Scan(&item_id);
		if resOp != nil {
			if pgErr, ok := err.(*pgconn.PgError); ok {
				fmt.Println(pgErr)
				return pgErr
			}
		}
	}
	return nil
}
// запись orderInfo делим на части: сама вставка полей order и отдельно вставка структур
func insertFullOrder(conn *pgx.Conn, data orderInfo) {
	resOp1 := insertOrderInfo(conn, data)
	if resOp1 != nil {
		fmt.Println("Ошибка вставки в таблицу orderInfo!")
	}
	resOp2 := insertPaymentInfo(conn, data)
	if resOp2 != nil {
		fmt.Println("Ошибка вставки в таблицу paymentInfo!")
	}
	resOp3 := insertDeliveryInfo(conn, data)
	if resOp3 != nil {
		fmt.Println("Ошибка вставки в таблицу deliveryInfo!")
	}
	resOp4 := insertItemInfo(conn, data)
	if resOp4 != nil {
		fmt.Println("Ошибка вставки в таблицу itemInfo!")
	}
	if resOp1 == nil && resOp2 == nil && resOp3 == nil && resOp4 == nil {
		fmt.Println("Данные удачно записаны в Postgres!")
	}
}
// по заданному uid ищем запись в БД и записываем в нашу структуру,
//вместо структуры сразу возвращаем json-сериализацию
func findOrder(conn *pgx.Conn, order_uid string) (string, error) {
	var resOrder orderInfo //результат
	var delivOrder deliveryInfo //часть результата - доставка
	var payOrder paymentInfo // платеж
	var itemsOrd [] itemInfo // массив товаров
	//сделаем несколько запросов - для каждой структуры отдельно
	sqlQueryStr1 := `SELECT * FROM orderInfo WHERE order_uid = $1`
	resOp1 := conn.QueryRow(context.Background(),sqlQueryStr1, order_uid).Scan(
	&resOrder.Order_uid, &resOrder.Track_number, &resOrder.Entry, &resOrder.Locale,
	&resOrder.Internal_signature, &resOrder.Customer_id, &resOrder.Delivery_service,
	&resOrder.Shardkey, &resOrder.Sm_id, &resOrder.Date_created, &resOrder.Oof_shard)
	if resOp1 != nil{
		if pgErr, ok := resOp1.(*pgconn.PgError); ok {
			fmt.Println("Ошибка в SELECT-запросе к orderInfo!")
			return "", pgErr
		}
	}
	//Мы получили orderInfo, однако его сложные поля (delivery,payment,items) еще надо заполнить

	//теперь записываем deliveryInfo
	sqlQueryStr2 := `SELECT name,phone,zip,city,address,region,email FROM deliveryInfo
	WHERE order_id = $1`
	resOp2 := conn.QueryRow(context.Background(), sqlQueryStr2, order_uid).Scan(
	&delivOrder.Name, &delivOrder.Phone, &delivOrder.Zip, &delivOrder.City,
	&delivOrder.Address, &delivOrder.Region, &delivOrder.Email)
	if resOp2 != nil{
		if pgErr, ok := resOp2.(*pgconn.PgError); ok {
			fmt.Println("Ошибка в SELECT-запросе к deliveryInfo!")
			return "", pgErr
		}
	}

	//получаем paymentInfo
	sqlQueryStr3 := `SELECT transaction,request_id,currency,provider,amount,payment_dt,bank,
	delivery_cost,goods_total,custom_fee FROM paymentInfo WHERE order_id = $1`
	resOp3 := conn.QueryRow(context.Background(), sqlQueryStr3, order_uid).Scan(
	&payOrder.Transaction, &payOrder.Request_id, &payOrder.Currency, &payOrder.Provider,
	&payOrder.Amount, &payOrder.Payment_dt, &payOrder.Bank, &payOrder.Delivery_cost,
	&payOrder.Goods_total, &payOrder.Custom_fee)
	if resOp3 != nil{
		if pgErr, ok := resOp3.(*pgconn.PgError); ok {
			fmt.Println("Ошибка в SELECT-запросе к paymentInfo!")
			return "", pgErr
		}
	}
	//остался items-слайс
	sqlQueryStr4 := `SELECT chrt_id, track_number, price, rid, name, sale, size, total_price,
	nm_id, brand, status FROM itemInfo WHERE order_id = $1`
	//теперь вместо queryRow используется просто query, потому что записей может быть больше одной
	itemRows, resOp4 := conn.Query(context.Background(), sqlQueryStr4, order_uid)
	if resOp4 != nil{
		if pgErr, ok := resOp4.(*pgconn.PgError); ok {
			fmt.Println("Ошибка в SELECT-запросе к itemInfo!")
			return "", pgErr
		}
	}
	//обходим все полученные строки из запроса
	for itemRows.Next(){
		var itemOrder itemInfo
		itemRows.Scan(&itemOrder.Chrt_id, &itemOrder.Track_number, &itemOrder.Price,
		&itemOrder.Rid, &itemOrder.Name, &itemOrder.Sale, &itemOrder.Size,
		&itemOrder.Total_price, &itemOrder.Nm_id, &itemOrder.Brand, &itemOrder.Status)
		itemsOrd = append(itemsOrd, itemOrder)//добавляем в слайс текущую строку
	}
	//дополняем orderInfo полученными структурами
	resOrder.Delivery=delivOrder
	resOrder.Payment=payOrder
	resOrder.Items=itemsOrd
	res, _ := json.Marshal(&resOrder)//получаем json-строку
	return string(res), nil
}
// Загрузить список всех uid'ов из базы данных
func loadUids(conn *pgx.Conn) (slice_uid []string) {
	sqlQueryStr := `SELECT array_agg(order_uid) from orderInfo`
	err := conn.QueryRow(context.Background(), sqlQueryStr).Scan(&slice_uid)
	if err != nil {
		fmt.Println("Ошибка в получении uid's из Postgres!")
	}
	return slice_uid
}

func main() {
	//загрузка .env
	envVars, err := initEnvironment()
	if err != nil {
		os.Exit(1)
	}
	// подключение к postgres
	connectionStr:= fmt.Sprintf("postgres://%s:%s@%s:%s/%s", envVars.username,
	envVars.password, envVars.host_name, envVars.postgres_port, envVars.db_name)
	connPSQL, err := pgx.Connect(context.Background(), connectionStr)
	if err != nil {
		fmt.Printf("Ошибка подключения к Postgres: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Удалось подключиться к Postgres!")
	defer connPSQL.Close(context.Background())

	//Загрузка массива всех uid'ов из базы для переноса в кэш
	uidSlice := loadUids(connPSQL)
	inMemCache := cache.New(-1, -1)
	for i := range uidSlice {
		data, err := findOrder(connPSQL, uidSlice[i])
		if err != nil {
			fmt.Println("Во время поиска по uid возникла ошибка!")
		}else{
		inMemCache.Set(uidSlice[i], data, cache.NoExpiration)}
	}
	cacheLen := len(inMemCache.Items())//кол-во записей
	fmt.Printf("Кэш загружен! Всего записей в кэше: %d\n", cacheLen)
	// Подключение к кластеру nats-streaming-service
	stanConn, err := stan.Connect("test-cluster", "simple-pub")
	if err != nil{
		fmt.Println("Не удалось подключиться к NATS-Streaming!")
		os.Exit(1)
	}
	defer stanConn.Close()
	var streaming_order orderInfo
	// подписка на канал для дальнейшей обработки полученных данных
	stanConn.Subscribe("service", func(m *stan.Msg) {
		resOp := json.Unmarshal(m.Data, &streaming_order)
		if resOp != nil {
			fmt.Println("NATS-STREAMING: получен некорректный JSON!")
		} else {
			fmt.Println("NATS-STREAMING: получен корректный JSON!")
			inMemCache.Set(streaming_order.Order_uid, string(m.Data), cache.NoExpiration)
			insertFullOrder(connPSQL, streaming_order)
			fmt.Println("Данные записаны в Postgres и кэш!")
		}
	})
	// http сервер который позволяет получить информацию о заказе по order_uid
	gin.SetMode(gin.ReleaseMode)
	routerGin := gin.Default()
	routerGin.LoadHTMLGlob("pages/*.html")
	routerGin.Static("/css","pages/css")
	routerGin.StaticFile("/apple-touch-icon.png","pages/favicons/apple-touch-icon.png")
	routerGin.StaticFile("/favicon-16x16.png","pages/favicons/favicon-16x16.png")
	routerGin.StaticFile("/favicon-32x32.png","pages/favicons/favicon-32x32.png")
	fmt.Println("HTML загружен")
	routerGin.GET("/", func(cont *gin.Context) {
		cont.HTML(http.StatusOK, "mainpage.html", gin.H{
			"content": "mainpage",
		})

	})
	routerGin.POST("/result", func(cont *gin.Context) {
		result, _ := inMemCache.Get(cont.PostForm("order_uid"))
		if result == nil {
			cont.PureJSON(http.StatusOK,"Не удалось получить запись по uid = "+cont.PostForm("order_uid"))

		} else {
			cont.PureJSON(http.StatusOK, result)
		}

	})
	HTTPConnStr:= fmt.Sprintf(":%s", envVars.http_port)
	fmt.Println("Запуск HTTP-сервера PORT:8080")
	routerGin.Run(HTTPConnStr)
}

