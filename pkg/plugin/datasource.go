package plugin

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana-plugin-sdk-go/live"
	"github.com/timeplus-io/proton-grafana-source/pkg/models"
	"github.com/timeplus-io/proton-grafana-source/pkg/timeplus"
)

var (
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ backend.StreamHandler         = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

func NewDatasource(ctx context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	logger := log.DefaultLogger.FromContext(ctx)
	conf, err := models.LoadPluginSettings(settings)
	if err != nil {
		return nil, err
	}

	engine := timeplus.NewEngine(logger, conf.Host, conf.TCPPort, conf.HTTPPort, conf.Username, conf.Secrets.Password)

	logger.Debug("new timeplus source created")

	return &Datasource{
		engine:  engine,
		queries: make(map[string]string),
	}, nil
}

// Datasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type Datasource struct {
	engine  timeplus.Engine
	queries map[string]string
}

func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	logger := log.DefaultLogger.FromContext(ctx)
	response := backend.NewQueryDataResponse()

	for _, query := range req.Queries {
		q := queryModel{}
		if err := json.Unmarshal(query.JSON, &q); err != nil {
			return nil, err
		}

		resp := backend.DataResponse{}
		frame := data.NewFrame("response")

		isStreaming, err := d.engine.IsStreamingQuery(ctx, q.SQL)
		if err != nil {
			return nil, err
		}

		if isStreaming {
			id := uuid.NewString()
			d.queries[id] = q.SQL
			channel := live.Channel{
				Scope:     live.ScopeDatasource,
				Namespace: req.PluginContext.DataSourceInstanceSettings.UID,
				Path:      id,
			}
			frame.SetMeta(&data.FrameMeta{Channel: channel.String()})
			resp.Frames = append(resp.Frames, frame)
		} else {
			count := 0
			columnTypes, ch, err := d.engine.RunQuery(ctx, q.SQL)
			if err != nil {
				return nil, err
			}

			for _, col := range columnTypes {
				frame.Fields = append(frame.Fields, timeplus.NewDataFieldByType(col.Name(), col.DatabaseTypeName()))
			}

		LOOP:
			for {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case row, ok := <-ch:
					if !ok {
						logger.Info("Query finished", "count", count)

						resp.Frames = append(resp.Frames, frame)
						break LOOP
					}

					fData := make([]any, len(columnTypes))
					for i, r := range row {
						col := columnTypes[i]
						fData[i] = timeplus.ParseValue(col.Name(), col.DatabaseTypeName(), nil, r, false)
						count++
					}

					frame.AppendRow(fData...)
				}
			}

		}

		response.Responses[query.RefID] = resp
	}

	return response, nil
}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSampleDatasource factory function.
func (d *Datasource) Dispose() {
	// log.DefaultLogger.Info("Dispose")
	if err := d.engine.Dispose(); err != nil {
		log.DefaultLogger.Error("failed to dispose", "error", err)
		return
	}
}

func (d *Datasource) SubscribeStream(ctx context.Context, req *backend.SubscribeStreamRequest) (*backend.SubscribeStreamResponse, error) {
	var status backend.SubscribeStreamStatus
	if _, ok := d.queries[req.Path]; ok {
		status = backend.SubscribeStreamStatusOK
	} else {
		status = backend.SubscribeStreamStatusNotFound
	}

	return &backend.SubscribeStreamResponse{
		Status: status,
	}, nil
}

func (d *Datasource) PublishStream(ctx context.Context, _ *backend.PublishStreamRequest) (*backend.PublishStreamResponse, error) {
	return &backend.PublishStreamResponse{
		Status: backend.PublishStreamStatusPermissionDenied,
	}, nil
}

func (d *Datasource) RunStream(ctx context.Context, req *backend.RunStreamRequest, sender *backend.StreamSender) error {
	logger := log.DefaultLogger.FromContext(ctx)

	path := req.Path
	sql, ok := d.queries[path]
	if !ok {
		return nil
	}

	columnTypes, ch, err := d.engine.RunQuery(ctx, sql)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case row, ok := <-ch:
			if !ok {
				logger.Warn("Streaming query terminated")
				return nil
			}
			frame := data.NewFrame("response")

			for _, c := range columnTypes {
				frame.Fields = append(frame.Fields, timeplus.NewDataFieldByType(c.Name(), c.DatabaseTypeName()))
			}

			fData := make([]any, len(columnTypes))
			for i, r := range row {
				col := columnTypes[i]
				fData[i] = timeplus.ParseValue(col.Name(), col.DatabaseTypeName(), nil, r, false)
			}

			frame.AppendRow(fData...)
			if err := sender.SendFrame(frame, data.IncludeAll); err != nil {
				logger.Error("Failed send frame", "error", err)
			}
		}
	}
}

type queryModel struct {
	SQL string `json:"sql"`
}

func (d *Datasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	logger := log.DefaultLogger.FromContext(ctx)
	res := &backend.CheckHealthResult{}
	config, err := models.LoadPluginSettings(*req.PluginContext.DataSourceInstanceSettings)

	if err != nil {
		res.Status = backend.HealthStatusError
		res.Message = "Unable to load settings"
		return res, nil
	}

	if len(config.Host) == 0 {
		res.Status = backend.HealthStatusError
		res.Message = "'Host' cannot be empty"
		return res, nil
	}

	engine := timeplus.NewEngine(logger, config.Host, config.TCPPort, config.HTTPPort, config.Username, config.Secrets.Password)

	if err := engine.Ping(ctx); err != nil {
		res.Status = backend.HealthStatusError
		res.Message = "failed to ping timeplusd: " + err.Error()
		return res, nil
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Proton data source is working",
	}, nil
}
