package executer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/chainreactors/neutron/common"
	"github.com/chainreactors/neutron/operators"
	"github.com/chainreactors/neutron/protocols"
)

type Executer struct {
	requests []protocols.Request
	options  *protocols.ExecuterOptions
}

type Event map[string]interface{}
type WrappedEvent struct {
	InternalEvent   Event
	OperatorsResult *operators.Result
}

var _ protocols.Executer = &Executer{}

// NewExecuter creates a new request executer for list of requests
func NewExecuter(requests []protocols.Request, options *protocols.ExecuterOptions) *Executer {
	return &Executer{requests: requests, options: options}
}

// Compile compiles the execution generators preparing any requests possible.
func (e *Executer) Compile() error {
	for _, request := range e.requests {
		err := request.Compile(e.options)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Executer) Options() *protocols.ExecuterOptions {
	return e.options
}

// Requests returns the total number of requests the rule will perform
func (e *Executer) Requests() int {
	var count int
	for _, request := range e.requests {
		count += request.Requests()
	}
	return count
}

// Execute executes the protocol group and returns true or false if results were found.
func (e *Executer) Execute(input *protocols.ScanContext) (*operators.Result, error) {
	var result *operators.Result

	previous := make(map[string]interface{})
	dynamicValues := common.MergeMaps(make(map[string]interface{}), input.Payloads)
	var events []protocols.InternalWrappedEvent
	for _, req := range e.requests {
		err := req.ExecuteWithResults(input, dynamicValues, previous, func(event *protocols.InternalWrappedEvent) {
			events = append(events, *event)
			lenevents := len(events)

			if event.OperatorsResult != nil {
				result = event.OperatorsResult
				// 初始化 Payloadreqresp map 如果为 nil
				if result.Payloadreqresp == nil {
					result.Payloadreqresp = make(map[string]interface{})
				}

				// 如果事件数量大于5，只取最后5个事件
				var eventsToProcess []protocols.InternalWrappedEvent
				if lenevents > 5 {
					eventsToProcess = events[lenevents-5:]
				} else {
					eventsToProcess = events
				}

				// 清空之前的请求数据，重新填充
				for i := 0; i < 5; i++ {
					delete(result.Payloadreqresp, "req"+strconv.Itoa(i))
				}

				// 只保存最后5个事件，命名为 req0-req4
				for i, wrappedEvent := range eventsToProcess {
					result.Payloadreqresp["req"+strconv.Itoa(i)] = ReconstructHTTPPacket(wrappedEvent.InternalEvent)
				}

				if len(result.Payloadreqresp) > 0 {
					result.Payloadreqresp["url"] = event.InternalEvent["host"]
				}
			}

		})
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

// HTTPPacketResult 包含还原后的完整请求和响应包，以及执行时间。
type HTTPPacketResult struct {
	RequestPacket  string
	ResponsePacket string
	ExecutionTime  string
}

// ReconstructHTTPPacket 接收 InternalEvent map 并还原 HTTP 请求和响应包。
// 它只依赖于你提供的特定字段。
func ReconstructHTTPPacket(event map[string]interface{}) HTTPPacketResult {
	executionTime := ""
	if event["duration"] != nil {
		executionTime = event["duration"].(string)
	}
	result := HTTPPacketResult{
		RequestPacket:  "Error: Request packet could not be reconstructed.",
		ResponsePacket: "Error: Response packet could not be reconstructed.",
		ExecutionTime:  executionTime,
	}

	// --- 1. 还原请求包 (Request Packet) ---

	// 1.1 提取请求头和请求体
	requestHeadersRaw, okReq := event["request"].(string)
	requestBody, okBody := event["reqbody"].(string)

	if okReq {
		// 清理头部字符串：
		// - 移除末尾可能的换行符或空格
		// - 替换头部中多余的 [ 和 ]，使其更像标准格式
		cleanedHeaders := strings.TrimSpace(requestHeadersRaw)
		cleanedHeaders = strings.ReplaceAll(cleanedHeaders, "[", "")
		cleanedHeaders = strings.ReplaceAll(cleanedHeaders, "]", "")

		// 拼接请求包
		requestBuilder := strings.Builder{}
		requestBuilder.WriteString(cleanedHeaders)

		if okBody && requestBody != "" {
			// 如果有 Body，添加空行分隔符
			requestBuilder.WriteString("\r\n\r\n")
			requestBuilder.WriteString(requestBody)
		} else {
			// 如果没有 Body 或 Body为空，仍然添加一个空行，完成头部
			requestBuilder.WriteString("\r\n\r\n")
		}

		result.RequestPacket = requestBuilder.String()
	}

	// --- 2. 还原响应包 (Response Packet) ---

	// 2.1 提取响应状态行和响应体
	responseStatusLine, okResp := event["response"].(string)
	responseBody, okRespBody := event["respbody"].(string)

	// 健壮性增强：如果不是 map[string][]string，则尝试 map[string]interface{}
	var headersMap map[string][]string

	if okResp {
		responseBuilder := strings.Builder{}
		responseStatusLine = strings.TrimSpace(responseStatusLine)
		responseStatusLine = strings.ReplaceAll(responseStatusLine, "[", "")
		responseStatusLine = strings.ReplaceAll(responseStatusLine, "]", "")
		// 状态行
		responseBuilder.WriteString(responseStatusLine)
		responseBuilder.WriteString("\r\n") // 添加 CRLF

		// 拼接头部
		if headersMap != nil {
			keys := make([]string, 0, len(headersMap))
			for k := range headersMap {
				keys = append(keys, k)
			}

			for _, key := range keys {
				values := headersMap[key]
				value := strings.Join(values, ", ")
				// 再次格式化为标准的 HTTP 头部 Key: Value
				responseBuilder.WriteString(fmt.Sprintf("%s: %s\r\n", strings.Title(strings.ToLower(key)), value))
			}
		}

		// 头部和正文分隔符
		responseBuilder.WriteString("\r\n")

		if okRespBody && responseBody != "" {
			// 响应体
			responseBuilder.WriteString(strings.TrimSpace(responseBody))
		}

		result.ResponsePacket = responseBuilder.String()
	}

	return result
}
