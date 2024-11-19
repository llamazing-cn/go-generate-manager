# Go Generate Manager

一个高效的 Go 代码生成工具管理器。

## 功能特点

- 并发执行代码生成命令
- 智能缓存，避免重复生成
- 支持多种代码生成工具 (mockgen, protoc 等)
- 自动检测文件变更
- 可配置的并发数量

## 安装

```bash
go install github.com/llamazing-cn/go-generate-manager/cmd/gogen@latest
```

## 使用方法

基本用法：
```bash
# 在当前目录执行 mockgen
gogen -c mockgen

# 指定目录
gogen -d ./pkg -c mockgen

# 指定输出目录
gogen -d ./pkg -c mockgen -o ./gen

# 自定义 worker 数量
gogen -c mockgen -w 4
```

完整参数说明：
```
Usage: gogen [options]

Options:
  -d, --dir     <path>     目标目录 (默认: 当前目录)
  -c, --cmd     <command>  生成命令 (必需)
  -o, --output  <path>     输出目录 (默认: 同源目录)
  -w, --workers <number>   worker数量 (默认: min(CPU数量*2, 8))
  -h, --help              显示帮助信息
```

## 性能基准测试

我们对不同数量的 worker 进行了基准测试，测试环境和结果如下：

### 测试环境
- **OS**: Darwin (macOS)
- **Arch**: arm64
- **CPU**: Apple M3
- **测试数据**: 20个包，每个包5个接口，每个接口约10个方法
- **总计**: 100个文件，约1000个需要生成的方法

### 测试结果

#### 执行时间
| Workers | 时间 (s) | 性能提升 |
| ------- | -------- | -------- |
| 1       | 4.53     | 基准线   |
| 2       | 2.85     | +37%     |
| 4       | 2.11     | +53%     |
| 8       | 2.05     | +54%     |
| 16      | 2.11     | +53%     |
| 32      | 2.15     | +52%     |
| 64      | 2.18     | +51%     |
| 128     | 2.21     | +51%     |

#### 内存使用
| Workers | 内存 (MB) | 分配次数 |
| ------- | --------- | -------- |
| 1       | 3.2       | 19,597   |
| 2       | 4.9       | 23,879   |
| 4       | 4.9       | 23,835   |
| 8       | 4.9       | 23,853   |
| 16      | 4.9       | 23,766   |
| 32      | 5.0       | 23,866   |
| 64      | 5.1       | 23,909   |
| 128     | 5.1       | 23,881   |

### 分析结论

1. **最佳 Worker 数量**: 8
   - 提供最快的执行时间 (2.05s)
   - 相比单 worker 提升了54%的性能
   - 内存使用保持在合理水平 (4.9MB)

2. **性能拐点**
   - 8个 workers 后性能开始下降
   - 16个以上的 workers 不会带来额外性能提升
   - 过多的 workers 反而会增加系统开销

3. **内存使用情况**
   - 2个 workers 时内存使用就达到稳定水平
   - 32个以上 workers 会导致内存使用缓慢增加
   - 内存分配次数在增加 workers 后保持稳定

4. **建议配置**
   - 小型项目 (<10 文件): 使用 4 个 workers
   - 中型项目 (10-50 文件): 使用 8 个 workers
   - 大型项目 (>50 文件): 仍建议使用 8 个 workers
   - 默认配置: `min(runtime.NumCPU() * 2, 8)`

### 测试代码

完整的基准测试代码位于 `tests/bench_test.go`。测试用例包括：
- 20个不同的包
- 每个包5个不同的接口文件
- 每个接口约10个方法
- 使用真实的 mockgen 进行代码生成

运行基准测试：
```bash
go test -bench=. -benchmem ./tests
```


