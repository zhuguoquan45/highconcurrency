# AI Agent 核心概念

## 什么是 AI Agent

AI Agent 是一个能够**感知环境、自主决策、调用工具、完成目标**的 AI 系统。与普通 LLM 对话不同，Agent 具备：

- **自主规划**：将复杂目标拆解为多步骤任务
- **工具调用**（Tool Use / Function Calling）：调用外部 API、数据库、代码执行等
- **记忆**：短期（对话上下文）+ 长期（向量数据库）
- **反思**：根据执行结果调整下一步行动

---

## Agent 核心架构

```
用户输入
   ↓
LLM（大脑）
   ↓
规划（Planning）→ 拆解任务
   ↓
工具调用（Tool Use）→ 执行动作
   ↓
观察结果（Observation）
   ↓
反思 / 继续规划（ReAct 循环）
   ↓
最终输出
```

---

## ReAct 模式（最主流）

**Reasoning + Acting** 交替进行：

```
Thought: 我需要查询今天的天气
Action: search_weather(city="北京")
Observation: 北京今天晴，25°C
Thought: 已获取天气，可以回答用户
Answer: 北京今天晴天，气温 25°C
```

每一轮：LLM 思考 → 选择工具 → 执行 → 观察结果 → 继续思考，直到任务完成。

---

## Function Calling / Tool Use

LLM 通过结构化输出告诉系统调用哪个函数、传什么参数。

### 定义工具

```go
tools := []Tool{
    {
        Name:        "get_weather",
        Description: "获取指定城市的天气信息",
        Parameters: map[string]any{
            "type": "object",
            "properties": map[string]any{
                "city": map[string]any{
                    "type":        "string",
                    "description": "城市名称",
                },
            },
            "required": []string{"city"},
        },
    },
}
```

### 执行流程

```
1. 用户: "北京今天天气怎么样？"
2. LLM 返回: { "tool": "get_weather", "args": {"city": "北京"} }
3. 系统执行 get_weather("北京") → "晴，25°C"
4. 将结果塞回上下文，LLM 生成最终回答
```

---

## 记忆系统

| 类型 | 实现 | 说明 |
|------|------|------|
| 短期记忆 | 对话上下文（Context Window） | 当前会话内有效 |
| 长期记忆 | 向量数据库（pgvector、Milvus、Weaviate） | 跨会话持久化 |
| 工作记忆 | 临时变量 / Scratchpad | 任务执行中间状态 |

### 向量检索（RAG）

```
用户问题 → Embedding 模型 → 向量
                              ↓
                    向量数据库相似度检索
                              ↓
                    召回相关文档片段
                              ↓
                    拼入 Prompt → LLM 回答
```

---

## 多 Agent 协作

### 常见模式

**串行（Pipeline）：**
```
Agent A（规划）→ Agent B（执行）→ Agent C（审核）
```

**并行（Parallel）：**
```
主 Agent → 子 Agent 1（搜索）
         → 子 Agent 2（计算）
         → 子 Agent 3（写作）
         → 汇总结果
```

**监督者模式（Supervisor）：**
```
Supervisor Agent
    ├── Worker Agent 1
    ├── Worker Agent 2
    └── Worker Agent 3
```

---

## 主流框架对比

| 框架 | 语言 | 特点 |
|------|------|------|
| LangChain | Python/JS | 生态最丰富，组件多 |
| LlamaIndex | Python | 专注 RAG 和数据索引 |
| AutoGen | Python | 微软出品，多 Agent 对话 |
| CrewAI | Python | 角色扮演式多 Agent |
| Dify | Python | 低代码，可视化编排 |
| Claude Agent SDK | Python | Anthropic 官方，工具调用强 |

---

## 面试高频问题

**Q: Agent 和普通 LLM 调用的区别？**
- 普通调用：一问一答，无状态
- Agent：多轮规划、工具调用、有记忆、自主决策

**Q: RAG 和 Fine-tuning 的区别？**
| 对比 | RAG | Fine-tuning |
|------|-----|-------------|
| 知识更新 | 实时（更新向量库）| 需重新训练 |
| 成本 | 低 | 高 |
| 适用 | 私有知识库问答 | 改变模型行为/风格 |
| 幻觉风险 | 低（有来源）| 较高 |

**Q: 如何防止 Agent 无限循环？**
- 设置最大迭代次数（max_iterations）
- 设置总超时时间（context.WithTimeout）
- 检测重复 Action，强制终止

**Q: Agent 的可靠性如何保证？**
- 工具调用结果校验
- 关键步骤人工确认（Human-in-the-loop）
- 回滚机制（对有副作用的操作）
- 日志记录每一步 Thought/Action/Observation
