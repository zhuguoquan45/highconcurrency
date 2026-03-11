# gRPC

## 什么是 gRPC

Google 开源的高性能 RPC 框架，基于 **HTTP/2 + Protobuf**，是微服务间通信的主流选择。

---

## 核心优势

| 特性 | 说明 |
|------|------|
| HTTP/2 | 多路复用、头部压缩、二进制帧，比 HTTP/1.1 高效 |
| Protobuf | 二进制序列化，比 JSON 小 3-10 倍，解析更快 |
| 强类型 | IDL 定义接口，自动生成代码，减少错误 |
| 流式通信 | 支持单向/双向流，适合实时场景 |
| 多语言 | 自动生成 Go/Java/Python 等客户端 |

---

## 四种通信模式

```protobuf
service OrderService {
    // 1. 一元 RPC（最常用）
    rpc GetOrder(OrderRequest) returns (OrderResponse);

    // 2. 服务端流
    rpc ListOrders(ListRequest) returns (stream OrderResponse);

    // 3. 客户端流
    rpc CreateOrders(stream OrderRequest) returns (CreateResponse);

    // 4. 双向流
    rpc Chat(stream Message) returns (stream Message);
}
```

---

## Protobuf 定义

```protobuf
syntax = "proto3";
package order;
option go_package = "./pb";

message OrderRequest {
    string order_id = 1;
    int64  user_id  = 2;
}

message OrderResponse {
    string order_id = 1;
    string status   = 2;
    double amount   = 3;
}
```

```bash
# 生成 Go 代码
protoc --go_out=. --go-grpc_out=. order.proto
```

---

## Go 实现示例

### Server
```go
type orderServer struct {
    pb.UnimplementedOrderServiceServer
}

func (s *orderServer) GetOrder(ctx context.Context, req *pb.OrderRequest) (*pb.OrderResponse, error) {
    // 检查 context 是否已取消
    if ctx.Err() != nil {
        return nil, status.Error(codes.Canceled, "request canceled")
    }
    return &pb.OrderResponse{OrderId: req.OrderId, Status: "paid"}, nil
}

func main() {
    lis, _ := net.Listen("tcp", ":50051")
    s := grpc.NewServer()
    pb.RegisterOrderServiceServer(s, &orderServer{})
    s.Serve(lis)
}
```

### Client
```go
conn, _ := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
defer conn.Close()

client := pb.NewOrderServiceClient(conn)
ctx, cancel := context.WithTimeout(context.Background(), time.Second)
defer cancel()

resp, err := client.GetOrder(ctx, &pb.OrderRequest{OrderId: "123"})
```

---

## 拦截器（Middleware）

```go
// 服务端拦截器（日志、认证、限流）
func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
    start := time.Now()
    resp, err := handler(ctx, req)
    log.Printf("method=%s duration=%v err=%v", info.FullMethod, time.Since(start), err)
    return resp, err
}

s := grpc.NewServer(grpc.UnaryInterceptor(loggingInterceptor))
```

---

## 错误处理

```go
import "google.golang.org/grpc/status"
import "google.golang.org/grpc/codes"

// 返回结构化错误
return nil, status.Errorf(codes.NotFound, "order %s not found", orderID)

// 客户端解析错误
if st, ok := status.FromError(err); ok {
    switch st.Code() {
    case codes.NotFound:
        // 处理 404
    case codes.DeadlineExceeded:
        // 处理超时
    }
}
```

---

## 面试高频问题

**Q: gRPC 和 HTTP REST 的区别？**
| 对比 | gRPC | REST |
|------|------|------|
| 协议 | HTTP/2 | HTTP/1.1 |
| 序列化 | Protobuf（二进制） | JSON（文本） |
| 性能 | 高 | 中 |
| 可读性 | 低 | 高 |
| 流式 | 支持 | 不支持 |
| 适用 | 内部微服务 | 对外 API |

**Q: gRPC 如何实现负载均衡？**
- 客户端负载均衡：`grpc.WithDefaultServiceConfig` 配置 round_robin
- 服务端负载均衡：通过 Nginx/Envoy 代理
- 结合服务发现（etcd/consul）动态更新地址列表

**Q: Protobuf 为什么比 JSON 快？**
- 二进制编码，不需要字段名，体积小
- 解析时直接按字段编号映射，不需要字符串解析
- 预编译的序列化代码，无反射开销

**Q: gRPC 如何做超时控制？**
- 通过 context.WithTimeout 传递截止时间
- gRPC 会自动将 deadline 通过 HTTP/2 header 传递到服务端
- 服务端可通过 ctx.Err() 检测是否超时
