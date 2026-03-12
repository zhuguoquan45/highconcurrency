# Agent 设计模式

## 1. ReAct（最常用）

Reasoning + Acting 循环，适合需要多步推理和工具调用的任务。

```
用户: 帮我分析 AAPL 股票并给出建议

Thought: 需要获取最新股价和财务数据
Action: get_stock_price(symbol="AAPL")
Observation: AAPL 当前价格 $185.2，今日涨幅 +1.2%

Thought: 还需要最近的新闻情绪
Action: search_news(query="Apple AAPL 2024")
Observation: 苹果发布新 AI 功能，市场反应积极

Thought: 数据已足够，可以给出分析
Answer: 基于当前价格和正面新闻...
```

**适用场景：** 信息检索、数据分析、代码生成

---

## 2. Plan-and-Execute

先整体规划，再逐步执行，适合复杂长任务。

```
用户: 帮我写一份竞品分析报告

[规划阶段]
Plan:
  1. 确定竞品列表
  2. 收集各竞品官网信息
  3. 分析功能对比
  4. 分析定价策略
  5. 撰写报告

[执行阶段]
Step 1: search("主要竞品") → [A, B, C]
Step 2: fetch_url(A官网), fetch_url(B官网), fetch_url(C官网)  ← 可并行
Step 3: analyze_features(...)
...
```

**适用场景：** 报告生成、项目规划、复杂研究任务

---

## 3. Reflection（反思模式）

Agent 生成结果后，由另一个 Agent（或自身）审查并改进。

```
Generator Agent → 初稿
      ↓
Critic Agent → 指出问题："第三段逻辑不清晰，缺少数据支撑"
      ↓
Generator Agent → 修改版
      ↓
Critic Agent → "通过" 或继续迭代
```

**适用场景：** 代码审查、文章写作、方案评审

---

## 4. Multi-Agent Supervisor

主 Agent 分配任务给专业子 Agent，适合需要多种专业能力的任务。

```
Supervisor
    ├── Research Agent（负责信息检索）
    ├── Code Agent（负责写代码）
    ├── Data Agent（负责数据分析）
    └── Writer Agent（负责撰写报告）
```

**Go 实现思路：**
```go
type Agent interface {
    Name() string
    Run(ctx context.Context, task string) (string, error)
}

type Supervisor struct {
    agents map[string]Agent
    llm    LLMClient
}

func (s *Supervisor) Dispatch(ctx context.Context, task string) (string, error) {
    // LLM 决定分配给哪个 Agent
    agentName, subtask := s.llm.Decide(task, s.agentNames())
    return s.agents[agentName].Run(ctx, subtask)
}
```

---

## 5. Human-in-the-Loop

关键步骤暂停，等待人工确认，适合高风险操作。

```go
func (a *Agent) executeAction(action Action) (string, error) {
    // 高风险操作需要人工确认
    if action.IsRisky() {
        confirmed := a.askHuman(fmt.Sprintf(
            "Agent 准备执行: %s\n参数: %v\n是否确认？(y/n)",
            action.Name, action.Args,
        ))
        if !confirmed {
            return "", errors.New("用户取消操作")
        }
    }
    return a.execute(action)
}
```

**适用场景：** 数据库写操作、发送邮件、支付、删除文件

---

## 6. Memory-Augmented Agent

结合长期记忆，让 Agent 记住用户偏好和历史上下文。

```go
type MemoryAgent struct {
    shortTerm []Message      // 当前对话
    longTerm  VectorStore    // 向量数据库
}

func (a *MemoryAgent) Chat(userMsg string) string {
    // 1. 从长期记忆检索相关历史
    memories := a.longTerm.Search(userMsg, topK=3)

    // 2. 构建带记忆的 Prompt
    systemPrompt := buildSystemPrompt(memories)

    // 3. 调用 LLM
    response := a.llm.Chat(systemPrompt, a.shortTerm, userMsg)

    // 4. 存入长期记忆
    a.longTerm.Store(userMsg, response)

    return response
}
```

---

## 选择哪种模式？

| 场景 | 推荐模式 |
|------|---------|
| 简单问答 + 工具调用 | ReAct |
| 复杂多步骤任务 | Plan-and-Execute |
| 需要高质量输出 | Reflection |
| 需要多种专业能力 | Multi-Agent Supervisor |
| 涉及高风险操作 | Human-in-the-Loop |
| 需要记住用户偏好 | Memory-Augmented |
