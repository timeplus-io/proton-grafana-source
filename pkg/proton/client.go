package proton

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/reactivex/rxgo/v2"
	protonGoDriver "github.com/timeplus-io/proton-go-driver/v2"
)

type Options struct {
	Host string `json:"host"`
	Port uint16 `json:"port"`
}

type Column struct {
	Name string
	Type string
}

type ProtonEngine struct {
	Options
	connection     *sql.DB
	runningQueries map[string]*ProtonQueryState //key is the query uuid
}

type ProtonQueryState struct {
	Query       string
	AddNow      bool
	Stream      chan rxgo.Item
	ColumnArray []Column
	Cancel      context.CancelFunc
}

func NewEngine(config Options) *ProtonEngine {
	if config.Host == "" {
		config.Host = "127.0.0.1"
	}

	if config.Port == 0 {
		config.Port = 8463
	}

	db := ProtonEngine{
		Options:        config,
		connection:     nil,
		runningQueries: make(map[string]*ProtonQueryState),
	}

	db.connect()

	return &db
}

func (e *ProtonEngine) connect() {
	connectionString := fmt.Sprintf("tcp://%s:%d?debug=%s", e.Options.Host, e.Options.Port, strconv.FormatBool(true)) //TODO: show debug info for now
	connect, err := sql.Open("proton", connectionString)
	if err != nil {
		log.DefaultLogger.Error("Fail to connect", err)
	} else {
		e.connection = connect
	}
}

func (e *ProtonEngine) StopQuery(id string) {
	if stat, exist := e.runningQueries[id]; exist {
		// do thread safe check
		stat.Cancel()
		if _, err := e.connection.Exec("kill query where query_id=@id", sql.Named("id", id)); err != nil {
			log.DefaultLogger.Error("failed to kill query", "queryID", id, "err", err)
		}
		log.DefaultLogger.Info("query stopped", "queryID", id)
	}
}

func (e *ProtonEngine) IsConnected() bool {
	if err := e.connection.Ping(); err != nil {
		log.DefaultLogger.Error("Fail to ping", err)
		return false
	}

	return true
}

func (e *ProtonEngine) GetQueryState(id string) ProtonQueryState {
	return *e.runningQueries[id]
}

func (e *ProtonEngine) RunQuery(sql string, id string, isStreaming bool, addNow bool) ([][]interface{}, error) {
	ctx, cancel := context.WithCancel(context.Background())
	ckCtx := protonGoDriver.Context(ctx, protonGoDriver.WithQueryID(id))

	rows, err := e.connection.QueryContext(ckCtx, sql)
	if err != nil {
		log.DefaultLogger.Error("[client.go] Failed to run query", "query", sql, "error", err.Error())
		cancel()
		return nil, err
	}

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		log.DefaultLogger.Error("Failed to get column type", "error", err.Error())
		cancel()
		return nil, err
	}
	count := len(columnTypes)
	header := make([]Column, count)

	cstream := make(chan rxgo.Item)

	// keep the query state here
	queryStat := ProtonQueryState{
		Query:       sql,
		AddNow:      addNow,
		Stream:      cstream,
		ColumnArray: header,
		Cancel:      cancel,
	}
	e.runningQueries[id] = &queryStat

	for i, col := range columnTypes {
		header[i] = Column{
			Name: col.Name(),
			Type: col.DatabaseTypeName(),
		}
		//log.DefaultLogger.Info("[client.go L129] Header is", "Name", col.Name(), "Type", col.DatabaseTypeName())
	}

	//log.DefaultLogger.Info("[client.go] Header is", "header", header, "isStreaming", isStreaming)
	if isStreaming {
		//now, start getting the data
		go func() {
			values := make([]interface{}, count) // values is raw data
			valuePtrs := make([]interface{}, count)

			for rows.Next() {
				row := make([]interface{}, count) // row is string data
				for i := range columnTypes {
					valuePtrs[i] = &values[i]
				}

				if err = rows.Scan(valuePtrs...); err != nil {
					//TODO avoid using c.stream
					cstream <- rxgo.Of(err)
					continue
				}

				//log.DefaultLogger.Info("scan result (outer)", "values", values)
				for i := range columnTypes {
					//var value interface{}
					rawValue := values[i]
					//log.DefaultLogger.Info("scanned result (inner)", "rawValue", rawValue)
					row[i] = rawValue
				}

				//log.DefaultLogger.Info("[client.go] Send row to query channel", "row", row)
				//TODO avoid using c.stream
				cstream <- rxgo.Of(row)

			}
			//TODO avoid using c.stream
			close(cstream)
			rows.Close()
		}()

		return nil, nil
	} else {
		//for non-streaming query, get all results (limit 1000) and return
		maxRows := 1000
		var results [][]interface{}
		values := make([]interface{}, count) // values is raw data
		valuePtrs := make([]interface{}, count)
		for rows.Next() {
			maxRows--
			if maxRows < 0 {
				log.DefaultLogger.Info("too many rows. Just return 1000")
				break
			}

			row := make([]interface{}, count) // row is string data
			for i := range columnTypes {
				valuePtrs[i] = &values[i]
			}

			if err = rows.Scan(valuePtrs...); err != nil {
				return nil, err
			}

			//log.DefaultLogger.Info("scan result (outer)", "values", values)
			for i := range columnTypes {
				//var value interface{}
				rawValue := values[i]
				//log.DefaultLogger.Info("scanned result (inner)", "rawValue", rawValue)
				row[i] = rawValue
			}

			//log.DefaultLogger.Info("send row to query channel", "row", row)
			results = append(results, row)
		}
		return results, nil
	}
}

func (e *ProtonEngine) IsSubscribed(path string) bool {
	return true
}

func (e *ProtonEngine) Subscribe(t string) {
}

func (e *ProtonEngine) Unsubscribe(t string) {
}

func (e *ProtonEngine) Dispose() {
	//this needs to be verified, never called
	log.DefaultLogger.Info("Client.Dispose")
}
