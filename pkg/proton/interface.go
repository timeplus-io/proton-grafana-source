package proton

type Client interface {
	RunQuery(query string, id string, isStreaming bool, addNow bool) ([][]interface{}, error)
	StopQuery(id string)
	IsStreamingQuery(query string) bool
	GetQueryState(id string) ProtonQueryState
	IsConnected() bool
	IsSubscribed(topic string) bool
	//Messages(topic string) ([]mqtt.Message, bool)
	Subscribe(topic string)
	Unsubscribe(topic string)
	Dispose()
}
