package main

///etc/my.cnf
//[mysqld]
//server_id = 1
//binlog_format = ROW
//binlog_row_image = FULL
//expire_logs_days = 10
//max_binlog_size = 10M
//log_bin = /var/log/mysql/mysql-bin.log

import (
	"fmt"
	"github.com/JackShadow/go-binlog-example/utils"
	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
	"github.com/roistat/go-clickhouse"
	log "log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func main() {
	transport := clickhouse.NewHttpTransport()
	chConnection := fmt.Sprintf("%s:%s", utils.GodotEnv("CLICKHOUSE_HOST"), utils.GodotEnv("CLICKHOUSE_PORT"))
	conn := clickhouse.NewConn(chConnection, transport)
	err := conn.Ping()
	if err != nil {
		panic(err)
	}
	log.Printf("ClickHouse Connected To: %s Successfull!", chConnection)
	//read ClickHouseSchema.sql file as string
	text, err := os.ReadFile("ClickHouseSchema.sql")
	if err != nil {
		panic(err)
	}
	//get every statement from text by ";"
	var statements []string
	statements = strings.Split(string(text), ";")
	//execute every statement
	for _, statement := range statements {
		//if empty statement - skip
		if statement == "" {
			continue
		}
		q := clickhouse.NewQuery(statement)
		q.Exec(conn)
	}

	//check migration table count
	chDatabase := utils.GodotEnv("CLICKHOUSE_DATABASE")
	query := clickhouse.NewQuery(fmt.Sprintf("SELECT * FROM %s.migration", chDatabase))
	iter := query.Iter(conn)
	var (
		id        string
		migration string
		batch     string
	)
	var lastId uint32
	for iter.Scan(&id, &migration, &batch) {
		i, _ := strconv.Atoi(batch)
		lastId = uint32(i)
	}
	if iter.Error() != nil {
		log.Panicln(iter.Error())
	}
	log.Printf("Last migration id: %d", lastId)
	if lastId == 0 {
		//copy data from mysql to clickhouse
		copyFromMysqlToClickHouse(conn, chDatabase)
	}

	// Define default dump configuration
	dumpCfg := canal.DumpConfig{
		ExecutionPath: "",
		Databases:     []string{"shoma"},
		Tables:        []string{"users"},
	}
	mysqlConnection := fmt.Sprintf("%s:%s", utils.GodotEnv("MYSQL_HOST"), utils.GodotEnv("MYSQL_PORT"))
	// Set up canal to connect to MySQL database
	cfg := canal.NewDefaultConfig()
	cfg.Addr = mysqlConnection
	cfg.User = utils.GodotEnv("MYSQL_USER")
	cfg.Password = utils.GodotEnv("MYSQL_PASSWORD")
	cfg.ServerID = 1
	cfg.Flavor = "mysql"
	cfg.HeartbeatPeriod = 10
	cfg.Dump = dumpCfg
	//cfg.Dump.Databases = []string{"shoma"}
	//cfg.Dump.TableDB = "shoma"
	//cfg.Dump.Tables = []string{"users"}
	// Create an instance of the eventHandler struct
	eventHandler := &eventHandler{
		*conn,
	}

	// Create a new Canal instance with the specified configuration and event handler
	can, err := canal.NewCanal(cfg)
	if err != nil {
		log.Fatalf("Failed to create Canal: %v", err)
	}

	// Register the event handler with the Canal instance
	can.SetEventHandler(eventHandler)

	// Start canal and subscribe to all binlog events
	err = can.Run()
	if err != nil {
		log.Fatalf("Failed to start Canal: %v", err)
	}

	// Wait for SIGINT or SIGTERM signals to stop the program
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	// Stop canal and exit program
	can.Close()
	log.Println("Canal stopped")
}

func copyFromMysqlToClickHouse(conn *clickhouse.Conn, chDatabase string) {
	query := clickhouse.NewQuery(fmt.Sprintf("SHOW TABLES FROM  %s;", chDatabase))
	log.Printf("query:%s", query)
	iter := query.Iter(conn)
	var (
		table string
	)
	mysqlConnection := fmt.Sprintf("%s:%s", utils.GodotEnv("MYSQL_HOST"), utils.GodotEnv("MYSQL_PORT"))
	mysqlDatabase := utils.GodotEnv("MYSQL_DATABASE")
	mysqlUsername := utils.GodotEnv("MYSQL_USER")
	mysqlPassword := utils.GodotEnv("MYSQL_PASSWORD")
	for iter.Scan(&table) {
		query = clickhouse.NewQuery(fmt.Sprintf("TRUNCATE TABLE %s.%s", chDatabase, table))
		query.Exec(conn)
		if table == "migration" {
			//get timestamp
			timestamp := time.Now().Unix()
			//insert migration table
			query = clickhouse.NewQuery(fmt.Sprintf("INSERT INTO %s.migration (id, migration, batch) VALUES (1, 'migration', %d);", chDatabase, timestamp))
		} else {
			query = clickhouse.NewQuery(fmt.Sprintf("INSERT INTO %s.%s SELECT * FROM mysql('%s', '%s', '%s', '%s', '%s');", chDatabase, table, mysqlConnection, mysqlDatabase, table, mysqlUsername, mysqlPassword))
		}
		log.Printf("query:%s", query)
		query.Iter(conn)
	}
	if iter.Error() != nil {
		log.Panicln(iter.Error())
	}
}

// Define a custom event handler to process binlog events
type eventHandler struct {
	clickhouse.Conn
}

func (h *eventHandler) String() string {
	//TODO implement me
	panic("implement me")
}

func (h *eventHandler) OnRotate(header *replication.EventHeader, r *replication.RotateEvent) error {
	// Do nothing
	return nil
}

func (h *eventHandler) OnTableChanged(header *replication.EventHeader, schema string, table string) error {
	// Do nothing
	return nil
}

func (h *eventHandler) OnDDL(header *replication.EventHeader, nextPos mysql.Position, queryEvent *replication.QueryEvent) error {
	// Print the DDL statement to the console
	log.Printf("DDL statement: %v", string(queryEvent.Query))

	return nil
}

func (h *eventHandler) OnRow(e *canal.RowsEvent) error {
	// Print the row event to the console
	//database name
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered in f %s", r)
		}
	}()

	chDatabase := utils.GodotEnv("CLICKHOUSE_DATABASE")

	if e.Table.Schema == chDatabase {
		mysqlConnection := fmt.Sprintf("%s:%s", utils.GodotEnv("MYSQL_HOST"), utils.GodotEnv("MYSQL_PORT"))
		mysqlDatabase := utils.GodotEnv("MYSQL_DATABASE")
		mysqlUsername := utils.GodotEnv("MYSQL_USER")
		mysqlPassword := utils.GodotEnv("MYSQL_PASSWORD")
		query := clickhouse.NewQuery(fmt.Sprintf("SELECT * FROM %s.migration", chDatabase))
		iter := query.Iter(&h.Conn)
		var (
			id        string
			migration string
			batch     string
		)
		var last_id uint32
		for iter.Scan(&id, &migration, &batch) {
			i, _ := strconv.Atoi(batch)
			last_id = uint32(i)
		}
		if iter.Error() != nil {
			log.Panicln(iter.Error())
		}
		if last_id <= e.Header.Timestamp {
			if e.Action == "update" {
				var i uint = 1
				for _, element := range e.Rows {
					//check i is odd
					if i%2 != 0 {
						i = i + 1
						continue
					}
					log.Printf("element: %v", element)
					var targetId = element[0]
					query = clickhouse.NewQuery(fmt.Sprintf("DELETE FROM  %s.%s WHERE id=%d;", chDatabase, e.Table.Name, targetId))
					log.Printf("query: %v", query)
					query.Iter(&h.Conn)
					query = clickhouse.NewQuery(fmt.Sprintf("INSERT INTO  %s.%s SELECT * FROM mysql('%s', '%s', '%s', '%s', '%s') where id=%d;", chDatabase, e.Table.Name, mysqlConnection, mysqlDatabase, e.Table.Name, mysqlUsername, mysqlPassword, targetId))
					log.Printf("query: %v", query)
					query.Iter(&h.Conn)
					query = clickhouse.NewQuery(fmt.Sprintf("DELETE FROM %s.migration WHERE id = 1;", chDatabase))
					log.Printf("query: %v", query)
					query.Iter(&h.Conn)
					query = clickhouse.NewQuery(fmt.Sprintf("INSERT INTO  %s.migration VALUES (1,'test', %d);", chDatabase, e.Header.Timestamp))
					log.Printf("query: %v", query)
					query.Iter(&h.Conn)
					log.Printf("Updated: %v", e.Table.Schema)
				}
			} else if e.Action == "delete" {
				for _, element := range e.Rows {
					var targetId = element[0]
					query = clickhouse.NewQuery(fmt.Sprintf("DELETE FROM  %s.%s WHERE id=%d;", chDatabase, e.Table.Name, targetId))
					log.Printf("query: %v", query)
					query.Exec(&h.Conn)
					query = clickhouse.NewQuery(fmt.Sprintf("DELETE FROM %s.migration WHERE id = 1;", chDatabase))
					log.Printf("query: %v", query)
					query.Iter(&h.Conn)
					query = clickhouse.NewQuery(fmt.Sprintf("INSERT INTO  %s.migration VALUES (1,'test',%d);", chDatabase, e.Header.Timestamp))
					query.Iter(&h.Conn)
					log.Printf("delete: %v", targetId)
				}
			} else if e.Action == "insert" {
				for _, element := range e.Rows {
					var targetId = element[0]
					query = clickhouse.NewQuery(fmt.Sprintf("DELETE FROM  %s.%s WHERE id=%d;", chDatabase, e.Table.Name, targetId))
					query = clickhouse.NewQuery(fmt.Sprintf("INSERT INTO  %s.%s SELECT * FROM mysql('%s', '%s', '%s', '%s', '%s') where id=%d;", chDatabase, e.Table.Name, mysqlConnection, mysqlDatabase, e.Table.Name, mysqlUsername, mysqlPassword, targetId))
					query.Iter(&h.Conn)
					query = clickhouse.NewQuery(fmt.Sprintf("DELETE FROM %s.migration WHERE id = 1;", chDatabase))
					log.Printf("query: %v", query)
					query.Iter(&h.Conn)
					query = clickhouse.NewQuery(fmt.Sprintf("INSERT INTO  %s.migration VALUES (1,'test',%d);", chDatabase, e.Header.Timestamp))
					query.Iter(&h.Conn)
					log.Printf("insert: %v", targetId)
				}
			}
			log.Printf("Database: %v", e.Table.Schema)
			log.Printf("Table: %v", e.Table)
			log.Printf("Log Id: %v", e.Header.LogPos)
			log.Printf("Action: %v", e.Action)
			log.Printf("Timestamp: %v", e.Header.Timestamp)
		}
	}

	return nil
}

func (h *eventHandler) OnGTID(x *replication.EventHeader, z mysql.GTIDSet) error {
	return nil
}

func (h *eventHandler) OnPosSynced(header *replication.EventHeader, pos mysql.Position, set mysql.GTIDSet, force bool) error {
	// Do nothing

	return nil
}

func (h *eventHandler) OnXID(*replication.EventHeader, mysql.Position) error {
	// Do nothing
	return nil
}

func (h *eventHandler) OnUnmarshal(data []byte) (interface{}, error) {
	// Do nothing
	return nil, nil
}

func (h *eventHandler) OnRawEvent(event *replication.BinlogEvent) error {
	// Do nothing
	return nil
}
