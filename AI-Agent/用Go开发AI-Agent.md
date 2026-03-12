# 用 Go 开发 AI Agent

## 调用 Claude API（Anthropic SDK）

```go
package main

import (
    "context"
    "fmt"
    "github.com/anthropics/anthropic-sdk-go"
    "github.com/anthropics/anthropic-sdk-go/option"
)

func main() {
    client := anthropic.NewClient(
        option.WithAPIKey("your-api-key"),
    )

    msg, err := client.Messages.New(context.Background(), anthropic.MessageNewParams{
        Model:     anthropic.F(anthropic.ModelClaude3_5SonnetLatest),
        MaxTokens: anthropic.F(int64(1024)),
        Messages: anthropic.F([]anthropic.MessageParam{
            anthropic.NewUserMessage(anthropic.NewTextBlock("你好，介绍一下自己")),
        }),
    })
    if err != nil {
        panic(err)
    }
    fmt.Println(msg.Content[0].Text)
}
```

---

## Tool Use（Function Calling）完整示例

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/anthropics/anthropic-sdk-go"
    "github.com/anthropics/anthropic-sdk-go/option"
)

// 定义工具
var tools = []anthropic.ToolParam{
    {
        Name:        anthropic.F("get_weather"),
        Description: anthropic.F("获取指定城市的实时天气"),
        InputSchema: anthropic.F[interface{}](map[string]any{
            "type": "object",
            "properties": map[string]any{
                "city": map[string]any{
                    "type":        "string",
                    "description": "城市名称，如：北京、上海",
                },
            },
            "required": []string{"city"},
        }),
    },
}

// 模拟工具执行
func executeTool(name string, input json.RawMessage) string {
    switch name {
    case "get_weather":
        var args struct{ City string `json:"city"` }
        json.Unmarshal(input, &args)
        return fmt.Sprintf("%s 今天晴天，气温 25°C，湿度 60%%", args.City)
    }
    return "unknown tool"
}

func runAgent(userMessage string) {
    client := anthropic.NewClient(option.WithAPIKey("your-api-key"))
    ctx := context.Background()

    messages := []anthropic.MessageParam{
        anthropic.NewUserMessage(anthropic.NewTextBlock(userMessage)),
    }

    for {
        resp, err := client.Messages.New(ctx, anthropic.MessageNewParams{
            Model:     anthropic.F(anthropic.ModelClaude3_5SonnetLatest),
            MaxTokens: anthropic.F(int64(1024)),
            Tools:     anthropic.F(tools),
            Messages:  anthropic.F(messages),
        })
        if err != nil {
            panic(err)
        }

        // 任务完成
        if resp.StopReason == anthropic.StopReasonEndTurn {
            for _, block := range resp.Content {
                if block.Type == anthropic.ContentBlockTypeText {
                    fmt.Println("Agent:", block.Text)
                }
            }
            break
        }

        // 需要调用工具
        if resp.StopReason == anthropic.StopReasonToolUse {
            // 将 assistant 回复加入上下文
            messages = append(messages, anthropic.NewAssistantMessage(resp.Content...))

            // 执行所有工具调用
            var toolResults []anthropic.ToolResultBlockParam
            for _, block := range resp.Content {
                if block.Type == anthropic.ContentBlockTypeToolUse {
                    result := executeTool(block.Name, block.Input)
                    fmt.Printf("调用工具 %s，结果: %s\n", block.Name, result)
                    toolResults = append(toolResults, anthropic.NewToolResultBlock(block.ID, result, false))
                }
            }

            // 将工具结果加入上下文
            messages = append(messages, anthropic.NewUserMessage(
                func() []anthropic.ContentBlockParamUnion {
                    var blocks []anthropic.ContentBlockParamUnion
                    for _, r := range toolResults {
                        blocks = append(blocks, r)
                    }
                    return blocks
                }()...,
            ))
        }
    }
}

func main() {
    runAgent("北京和上海今天天气怎么样？帮我对比一下")
}
```

---

## 流式输出（Streaming）

```go
stream := client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
    Model:     anthropic.F(anthropic.ModelClaude3_5SonnetLatest),
    MaxTokens: anthropic.F(int64(1024)),
    Messages:  anthropic.F(messages),
})

for stream.Next() {
    event := stream.Current()
    switch delta := event.Delta.(type) {
    case anthropic.ContentBlockDeltaEventDelta:
        if delta.Type == anthropic.ContentBlockDeltaEventDeltaTypeTextDelta {
            fmt.Print(delta.Text) // 实时打印
        }
    }
}
if err := stream.Err(); err != nil {
    panic(err)
}
```

---

## RAG 实现（结合 pgvector）

```go
// 1. 文档向量化存储
func storeDocument(db *pgxpool.Pool, content string, embedding []float32) error {
    _, err := db.Exec(ctx,
        `INSERT INTO documents (content, embedding) VALUES ($1, $2)`,
        content, pgvector.NewVector(embedding),
    )
    return err
}

// 2. 相似度检索
func searchSimilar(db *pgxpool.Pool, queryEmbedding []float32, topK int) ([]string, error) {
    rows, err := db.Query(ctx,
        `SELECT content FROM documents
         ORDER BY embedding <=> $1  -- cosine distance
         LIMIT $2`,
        pgvector.NewVector(queryEmbedding), topK,
    )
    // ...
}

// 3. 拼入 Prompt
func buildRAGPrompt(question string, docs []string) string {
    context := strings.Join(docs, "\n\n")
    return fmt.Sprintf(`根据以下资料回答问题：

%s

问题：%s`, context, question)
}
```

---

## Agent 并发执行多工具

```go
// 并行执行多个工具调用
func executeToolsConcurrently(toolCalls []ToolCall) []ToolResult {
    results := make([]ToolResult, len(toolCalls))
    var wg sync.WaitGroup

    for i, call := range toolCalls {
        wg.Add(1)
        go func(idx int, tc ToolCall) {
            defer wg.Done()
            results[idx] = ToolResult{
                ID:     tc.ID,
                Output: executeTool(tc.Name, tc.Input),
            }
        }(i, call)
    }

    wg.Wait()
    return results
}
```

---

## 关键工程实践

### 超时控制
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
// 传入所有 API 调用
```

### 重试机制
```go
// 指数退避重试
for attempt := 0; attempt < 3; attempt++ {
    resp, err := client.Messages.New(ctx, params)
    if err == nil {
        break
    }
    time.Sleep(time.Duration(1<<attempt) * time.Second)
}
```

### 成本控制
- 记录每次调用的 input_tokens / output_tokens
- 设置 max_tokens 上限
- 缓存相同问题的回答（Redis）

### 安全
- Prompt Injection 防护：不直接将用户输入拼入系统 Prompt
- 工具权限最小化：只暴露必要的工具
- 敏感操作需二次确认（Human-in-the-loop）
