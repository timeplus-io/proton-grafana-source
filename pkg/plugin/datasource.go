package plugin

import (
	"context"
	"encoding/json"
	"time"

	"timeplus-proton-datasource/pkg/parser"
	"timeplus-proton-datasource/pkg/proton"

	"github.com/google/uuid"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana-plugin-sdk-go/live"
)

// Make sure Datasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler interfaces. Plugin should not implement all these
// interfaces- only those which are required for a particular task.
var (
	_ backend.QueryDataHandler      = (*ProtonDatasource)(nil)
	_ backend.CheckHealthHandler    = (*ProtonDatasource)(nil)
	_ instancemgmt.InstanceDisposer = (*ProtonDatasource)(nil)
)

func getDatasourceSettings(s backend.DataSourceInstanceSettings) (*proton.Options, error) {
	settings := &proton.Options{}
	if err := json.Unmarshal(s.JSONData, settings); err != nil {
		return nil, err
	}
	return settings, nil
}

type ProtonDatasource struct {
	logger log.Logger
	client proton.Client
}

// NewDatasource creates a new datasource instance.
func NewDatasource(ctx context.Context, s backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	settings, err := getDatasourceSettings(s)
	if err != nil {
		return nil, err
	}
	logger := log.NewWithLevel(log.Info)
	//logger.Info("NewProtonDatasource called", "settings", settings)

	client := proton.NewEngine(*settings)

	return &ProtonDatasource{
		logger,
		client}, nil
}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSampleDatasource factory function.
func (d *ProtonDatasource) Dispose() {
	// d.logger.Info("[plugin.go] Dispose called")
	// Clean up datasource instance resources.
	d.client.Dispose()
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *ProtonDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := d.query(ctx, req.PluginContext, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

type queryModel struct {
	AddNow      bool   `json:"addNow"`
	IsStreaming bool   `json:"isStreaming"`
	Query       string `json:"queryText"`
}

func (d *ProtonDatasource) query(_ context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	response := backend.DataResponse{}

	// Unmarshal the JSON into our queryModel.
	var qm queryModel

	response.Error = json.Unmarshal(query.JSON, &qm)
	if response.Error != nil {
		return response
	}
	//qm.Query can be null or empty String. Need to skip query to avoid error
	if qm.Query == "" {
		// d.logger.Info("[plugin.go] skip running the empty query")
		return response
	}
	// Generate an UUID for the proton query
	id := uuid.Must(uuid.NewRandom()).String()
	// d.logger.Info("[plugin.go] query with", "SQL", qm.Query, "QueryID", id, "RefID", query.RefID)

	rows, err := d.client.RunQuery(qm.Query, id, qm.IsStreaming, qm.AddNow)
	if err != nil {
		response.Error = err
		d.logger.Error("[plugin.go] client.RunQuery failed. Cannot submit the query.", "error", err)
		return response
	}

	// create data frame response
	frame := data.NewFrame("response")
	lenOfNow := 0
	if qm.AddNow {
		//if AddNow is one, we add the first column as now()
		frame.Fields = append(frame.Fields, parser.NewTimeField("time", false))
		lenOfNow = 1
	}

	for _, c := range d.client.GetQueryState(id).ColumnArray {
		frame.Fields = append(frame.Fields, parser.NewDataFieldByType(c.Name, c.Type))
	}

	if qm.IsStreaming {
		// to subscribe on a client-side and consume updates from a plugin.
		channel := live.Channel{
			Scope:     live.ScopeDatasource,
			Namespace: pCtx.DataSourceInstanceSettings.UID,
			Path:      id, //important! use query id as the path, so that RunStream() can get it via req.Path
		}
		frame.SetMeta(&data.FrameMeta{Channel: channel.String()})
	} else {
		//if it's not a streaming quer, then show all results in response
		for _, row := range rows {
			currentRow := make([]interface{}, len(row)+lenOfNow)
			if qm.AddNow {
				currentRow[0] = time.Now()
			}
			for i, r := range row {
				currentRow[i+lenOfNow] = parser.ParseValue("whatever", d.client.GetQueryState(id).ColumnArray[i].Type, nil, r, false)
			}
			frame.AppendRow(currentRow...)
		}
	}
	// add the frames to the response.
	response.Frames = append(response.Frames, frame)
	return response
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *ProtonDatasource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	//d.logger.Info("CheckHealth called", "request", req)

	if !d.client.IsConnected() {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "proton Disconnected",
		}, nil
	}

	var status = backend.HealthStatusOk
	var message = "Connnected to proton"

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}

func (d *ProtonDatasource) SubscribeStream(ctx context.Context, req *backend.SubscribeStreamRequest) (*backend.SubscribeStreamResponse, error) {
	// d.logger.Info("SubscribeStream %v", req)

	return &backend.SubscribeStreamResponse{
		Status: backend.SubscribeStreamStatusOK,
	}, nil
}

func (d *ProtonDatasource) PublishStream(ctx context.Context, req *backend.PublishStreamRequest) (*backend.PublishStreamResponse, error) {
	// d.logger.Info("PublishStream called", "request", req)

	// Do not allow publishing at all.
	return &backend.PublishStreamResponse{
		Status: backend.PublishStreamStatusPermissionDenied,
	}, nil
}

func (d *ProtonDatasource) RunStream(ctx context.Context, req *backend.RunStreamRequest, sender *backend.StreamSender) error {
	state := d.client.GetQueryState(req.Path)

	// Stream data frames periodically till stream closed by Grafana.
	for {
		select {
		case <-ctx.Done():
			//TODO, sometimes the 2nd streaming chart will get cancelled somehow
			// d.logger.Info("[plugin.go] Context canceled, finish streaming", "path", req.Path)
			d.client.StopQuery(req.Path)
			return nil
		case item := <-state.Stream:
			frame := data.NewFrame("response")
			lenOfNow := 0
			if state.AddNow {
				//if AddNow is one, we add the first column as now()
				frame.Fields = append(frame.Fields, parser.NewTimeField("time", false))
				lenOfNow = 1
			}

			for _, c := range state.ColumnArray {
				frame.Fields = append(frame.Fields, parser.NewDataFieldByType(c.Name, c.Type))
			}
			row := item.V.([]interface{})
			currentRow := make([]interface{}, len(row)+lenOfNow)
			if state.AddNow {
				currentRow[0] = time.Now()
			}
			for i, r := range row {
				currentRow[i+lenOfNow] = parser.ParseValue("whatever", state.ColumnArray[i].Type, nil, r, false)
			}
			frame.AppendRow(currentRow...)

			err := sender.SendFrame(frame, data.IncludeAll)
			if err != nil {
				d.logger.Error("Error sending frame", "error", err)
				continue
			}
		}
	}
}
