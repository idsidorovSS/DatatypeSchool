package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	_ "github.com/lib/pq"
)

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("Staring web listener in jsonschool...")
	log.Fatal(http.ListenAndServe(":8083", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("\n\nGot new request from IP %v \nTime= %v;  Method= %v;  Payload size= %v bytes",
		r.RemoteAddr, time.Now().Format(time.RFC3339), r.Method, r.ContentLength)
	fmt.Printf("\nChecking headers...")
	if r.Header.Get("Content-Type") != "application/json" {
		fmt.Printf("\ntoo bad, request Content-Type is no JSON! " +
			" Processing will not!")
	} else {
		fmt.Printf(" OK!\nChecking sender...")
		if strings.Contains(string(r.RemoteAddr), "192.168.1.47:") {
			fmt.Printf(" OK!\nChecking JSON...")
			var v map[string]interface{}
			json.NewDecoder(r.Body).Decode(&v)
			lora_DevEUI_uplink, _ := v["DevEUI_uplink"].(map[string]interface{})
			/**/ fmt.Printf("\n%s", lora_DevEUI_uplink)
			if len(lora_DevEUI_uplink) == 0 {
				fmt.Printf("\ntoo bad, key 'DevEUI_uplinк' not found! " +
					" Processing will not!")
			} else {
				fmt.Printf(" OK!")
				go write_to_base(v) // GOROUTINE !!!
			}
		} else {
			fmt.Printf("\ntoo bad, sender unkown! Processing will not!")
		}
	}
}

//////////////////////////////////////////////////////////
func write_to_base(v map[string]interface{}) {
	const (
		host     = "localhost"
		port     = 5432
		user     = "***"
		password = "***"
		dbname   = "***"
		tblname  = "***"
	)
	lora_DevEUI_uplink, _ := v["DevEUI_uplink"].(map[string]interface{})

	str_lora_Time := fmt.Sprintf("%v", lora_DevEUI_uplink["Time"])
	/**/ fmt.Println(str_lora_Time)
	str_lora_DevEUI := fmt.Sprintf("%v", lora_DevEUI_uplink["DevEUI"])
	str_lora_Fport := fmt.Sprintf("%v", lora_DevEUI_uplink["FPort"])
	str_lora_payload_hex := fmt.Sprintf("%v", lora_DevEUI_uplink["payload_hex"])
	lora_rawJson, _ := json.Marshal(v)

	const (
		needField = "loratime, loradeveui, lorafport, lorapayload_nex, rawlorajson"
	) //   Остановился здесь!
	//проверка, что номер порта число и не более 4 разрядов
	str_lora_Fport = "9999"
	var lora_Fport = 0
	portIsgood := true
	for i, element := range str_lora_Fport {
		fmt.Println(string(element))
		if i > 3 {
			fmt.Println("слишком длинное число!")
			portIsgood = false
			break
		} else if !unicode.IsDigit(element) {
			fmt.Println("это не число")
			portIsgood = false
			break
		}
	}
	if portIsgood {
		fmt.Println("конвертируем")
		lora_Fport, _ = strconv.Atoi(str_lora_Fport)
	}
	fmt.Println("пишем в базу ", lora_Fport)
	//if err != nil {
	//	log.Fatal(err)
	//	fmt.Printf("это не число!")
	//}

	needValues := fmt.Sprintf("'%s','%s',%v,'%s', '%s'",
		str_lora_Time,
		str_lora_DevEUI,
		lora_Fport, ///////////////заняться приведением типов данных//////////////////
		str_lora_payload_hex,
		lora_rawJson)
	fmt.Printf("\nneedValues=  " + needValues + "\n")

	sqlExecStr := fmt.Sprintf("insert into %s (%s) values (%s);", tblname, needField, needValues)
	// fmt.Printf("\nsqlExecStr=  " + sqlExecStr + "\n")

	dbConnStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		log.Fatal(err)
		fmt.Printf("ошибка sql.Open =", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err)
		fmt.Printf("ошибка db/Ping =", err)
	}
	defer db.Close()
	fmt.Println("Write_to_base...")
	result, err := db.Exec(sqlExecStr)
	if err != nil {
		panic(err)
	}
	fmt.Println(result.RowsAffected())
	fmt.Printf("Ok!")
}
