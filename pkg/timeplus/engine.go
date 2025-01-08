package timeplus

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"

	protonDriver "github.com/timeplus-io/proton-go-driver/v2"
)

const (
	bufferSize     = 1000
	defaultTimeout = 10 * time.Second
)

type Column struct {
	Name string
	Type string
}

type TimeplusEngine struct {
	connection *sql.DB
	logger     log.Logger
	analyzeURL string
	pingURL    string
	client     *http.Client
	header     http.Header
}

func NewEngine(logger log.Logger, host string, tcpPort, httpPort int, username, password string) *TimeplusEngine {
	connection := protonDriver.OpenDB(&protonDriver.Options{
		Addr: []string{fmt.Sprintf("%s:%d", host, tcpPort)},
		Auth: protonDriver.Auth{
			Username: username,
			Password: password,
		},
		DialTimeout: defaultTimeout,
		Debug:       false,
	})

	header := http.Header{}
	header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password))))
	header.Set("Content-Type", "application/json")

	return &TimeplusEngine{
		connection: connection,
		logger:     logger,
		analyzeURL: fmt.Sprintf("http://%s:%d/proton/v1/sqlanalyzer", host, httpPort),
		pingURL:    fmt.Sprintf("http://%s:%d/proton/ping", host, httpPort),
		header:     header,
		client: &http.Client{
			Timeout: defaultTimeout,
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout: defaultTimeout,
				}).Dial,
				TLSHandshakeTimeout: defaultTimeout,
			},
		},
	}
}

func (e *TimeplusEngine) Ping(ctx context.Context) error {
	if err := e.pingHttp(ctx); err != nil {
		return fmt.Errorf("failed to ping via http: %w", err)
	}

	if err := e.connection.Ping(); err != nil {
		return fmt.Errorf("failed to ping via tcp: %w", err)
	}

	return nil
}

func (e *TimeplusEngine) RunQuery(ctx context.Context, sql string) ([]*sql.ColumnType, chan []any, error) {
	ckCtx := protonDriver.Context(ctx)

	rows, err := e.connection.QueryContext(ckCtx, sql)
	if err != nil {
		return nil, nil, err
	}

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, nil, err
	}

	ch := make(chan []any, bufferSize)

	go func() {
		defer func() {
			close(ch)
		}()

		count := len(columnTypes)
		for rows.Next() {
			values := make([]any, count) // values is raw data
			valuePtrs := make([]any, count)

			row := make([]any, count) // row is string data
			for i := range columnTypes {
				valuePtrs[i] = &values[i]
			}
			if err = rows.Scan(valuePtrs...); err != nil {
				return
			}
			for i := range columnTypes {
				rawValue := values[i]
				row[i] = rawValue
			}

			ch <- row
		}

		rows.Close()
	}()

	return columnTypes, ch, nil
}

func (e *TimeplusEngine) Dispose() error {
	return e.connection.Close()
}

func (e *TimeplusEngine) IsStreamingQuery(ctx context.Context, query string) (bool, error) {
	queryMap := map[string]string{"query": query}
	jsonData, err := json.Marshal(queryMap)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.analyzeURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return false, err
	}
	req.Header = e.header

	resp, err := e.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 399 {
		var errStr string

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			errStr = err.Error()
		} else {
			errStr = string(body)
		}

		return false, fmt.Errorf("failed to analyze code: %d, error: %s", resp.StatusCode, errStr)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	var response map[string]interface{}
	if err = json.Unmarshal(body, &response); err != nil {
		return false, err
	}

	isStreaming, ok := response["is_streaming"].(bool)
	if !ok {
		return false, fmt.Errorf("invalid response %s", response)
	}

	return isStreaming, nil
}

func (e *TimeplusEngine) pingHttp(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, e.pingURL, nil)
	if err != nil {
		return err
	}
	req.Header = e.header

	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to ping, got %d", resp.StatusCode)
	}

	return nil
}
