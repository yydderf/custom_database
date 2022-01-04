package database

import (
    "time"
	"database/sql"
	"log"
    "fmt"

    "github.com/gocolly/colly"
    _ "github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func GetTime() int64 {
    return time.Now().Unix()
}

func scrapeUrl(isin string) string {
	return "https://tw.stock.yahoo.com/" + isin
}

type Stock struct{
    Name       string `json:"name"`
    Symbol     string `json:"symbol"`    //代號
    Price      string `json:"price"`
    Change     string `json:"change"`
    Changep    string `json:"changep"`   //change percentage 漲跌幅
    Buyin      string `json:"buyin"`
    Sellout    string `json:"sellout"`
    Opening    string `json:"opening"`
    Closing    string `json:"closing"`
    Highest    string `json:"highest"`
    Lowest     string `json:"lowest"`
    Volume     string `json:"volume"`
    Time       string `json:"time"`
  }

//temporary
func printStock(s Stock){
    fmt.Printf("%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s\n",s.Symbol,s.Name,s.Price,s.Change,s.Changep,s.Buyin,s.Sellout,s.Opening,s.Closing,s.Highest,s.Lowest,s.Volume,s.Time)
}

func GetData(db *sql.DB, time_t []string, i int) ([]Stock) {
    fmt.Printf("%v\n", time_t[i])
    var stockdata []Stock
    var s Stock
    rows, err := db.Query("SELECT * FROM epoch" + time_t[i])
    if err != nil {
        log.Fatalf("%q", err)
        return stockdata
    }

    defer rows.Close()
    for rows.Next() {
        if err := rows.Scan(&s.Name,&s.Symbol,&s.Price,&s.Change,&s.Changep,&s.Buyin,&s.Sellout,&s.Opening,&s.Closing,&s.Highest,&s.Lowest,&s.Volume,&s.Time); err != nil {
            log.Fatalf("%q", err)
            return stockdata
        }
        stockdata = append(stockdata, s)
    }
    if err = rows.Err(); err != nil {
        log.Fatalf("%q", err)
        return stockdata
    }
    return stockdata
}

//get datas , return type: struct array 
func Getdatas(db *sql.DB, time_t []string) (map[string][]Stock){
    var stockdataMap map[string][]Stock
    for i := range time_t {
        stockdata := GetData(db, time_t, i)
		if len(stockdata) == 0 {
			fmt.Printf("%s\n", "no data found")
            return stockdataMap
		} else {
            for _,s := range stockdata{
                printStock(s)
            }
		}
        stockdataMap[time_t[i]] = stockdata
	}
    return stockdataMap
}

func adddata(db *sql.DB,s Stock, n string) {
		_,err := db.Exec("INSERT INTO epoch"+n+" (name,symbol,price,change,changep,buyin,sellout,opening,closing,highest,lowest,volume,time) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) ",s.Name,s.Symbol,s.Price,s.Change,s.Changep,s.Buyin,s.Sellout,s.Opening,s.Closing,s.Highest,s.Lowest,s.Volume,s.Time)
        if err != nil {
            panic(err)
        }
}


func SetTable(db *sql.DB, tableName string) {
	if _, err := db.Exec("CREATE TABLE epoch" + tableName + " (name VARCHAR(100),symbol VARCHAR(100), price VARCHAR(100), change VARCHAR(100), changep  VARCHAR(100), buyin VARCHAR(100),sellout  VARCHAR(100), opening  VARCHAR(100), closing  VARCHAR(100), highest  VARCHAR(100), lowest  VARCHAR(100),volume  VARCHAR(100),time VARCHAR(100))"); err != nil {
	    return
	}
}


func Setdata(db *sql.DB,types []string,tableName string){
    c := colly.NewCollector()
    c.OnHTML(".table-body-wrapper > ul > li",func(h *colly.HTMLElement){
        tmp := make([]string,0,0)
        h.ForEach(".Fxg\\(1\\) > .Jc\\(fe\\)", func(i int, h *colly.HTMLElement){
                tmp = append(tmp, h.Text)
        })
        flag:=0
        h.ForEach(".Fxg\\(1\\) > .C\\(\\$c-trend-up\\)", func(i int, h *colly.HTMLElement){
                flag=1
        })
        if flag==0 {
            tmp[1] = "-" + tmp[1]
            tmp[2] = "-" + tmp[2]
        }else {
            tmp[1] = "+" + tmp[1]
            tmp[2] = "+" + tmp[2]
        }
        var s Stock
        s.Name   = h.ChildText(".Lh\\(20px\\).Fw\\(600\\).Fz\\(16px\\).Ell")
        s.Symbol = h.ChildText(".D\\(f\\).Ai\\(c\\) >  span ")
        s.Price  = tmp[0]
        s.Change = tmp[1]
        s.Changep= tmp[2]
        s.Buyin  = tmp[3]
        s.Sellout= tmp[4]
        s.Opening= tmp[5]
        s.Closing= tmp[6]
        s.Highest= tmp[7]
        s.Lowest = tmp[8]
        s.Volume = tmp[9]
        s.Time   = h.ChildText(".Fxs\\(1\\) > .Fz\\(14px\\)")
        //fmt.Println(s.Name)
        adddata(db,s,tableName)
    })
        for _, i := range types {
		    c.Visit(scrapeUrl(i))
	    }
        c.Wait()
}

func DropEpochTables(db *sql.DB) {
    log.Println("Dropping all tables from previous sessions")
    rows, err := db.Query("SELECT tablename FROM pg_tables WHERE tablename LIKE 'epoch%'")
    if err != nil {
        log.Fatalf("Failed to get tablenames from psql, %q", err)
    }
    defer rows.Close()
    var tablename string
    for rows.Next() {
        rows.Scan(&tablename)
        _, err = db.Exec("DROP TABLE IF EXISTS " + tablename);
        if err != nil {
            log.Fatalf("Failed to drop table: %s, %q", tablename, err)
        }
    }
}

func (st Stock) InfoPrice() string {
    return fmt.Sprintf("%-9s: %s\n", st.Symbol, st.Price)
}

func (st Stock) InfoExpPrice() string {
    return fmt.Sprintf("%s(%s)\n: %s\n", st.Name, st.Symbol, st.Price)
}

func (st Stock) InfoChangeP() string {
    return fmt.Sprintf("%-9s: %s\n", st.Symbol, st.Changep)
}

func (st Stock) InfoExpChangeP() string {
    return fmt.Sprintf("%s(%s)\n: %s\n", st.Name, st.Symbol, st.Changep)
}
