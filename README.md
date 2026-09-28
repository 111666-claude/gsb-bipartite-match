# bipartite-match

二分图最大匹配与最小顶点覆盖（König 定理），只用 Go 标准库。

```
go test ./...
go run ./cmd/bipartite-match --edges 'l1-r1,l1-r2,l2-r1'
go run ./cmd/bipartite-match --edges 'l1-r1,l1-r2,l2-r1' --cover
```

## 口径（README 为准）

- **要真的找最大匹配**：`--edges 'l1-r1,l1-r2,l2-r1'` 按口径是 `matching=2`、`pairs=l1-r2,l2-r1`，
  现在按边的书写顺序给 l1 贪心挑 r1，给 `matching=1`。
- **最小顶点覆盖**：`--cover` 输出覆盖所有边的最小顶点集合（按字典序），大小必须等于最大匹配；
  现在 `--cover` 直接报「最小顶点覆盖还没实现」。
- 输出确定：`pairs` 与 `cover` 都按字典序，同一输入重复运行结果相同。
- 输入：`--edges` 是逗号分隔的 `左-右`；空顶点名、缺 `-`、重复边都要报错并指出是第几条。
- 不变量：配对的左右顶点两两不同；匹配大小 ≤ 两侧顶点数；`cover` 必须覆盖每一条边；`|cover| == matching`。
- 规模：两侧各 1 万点、10 万条边要在 3 秒内出结果（用 Hopcroft-Karp 的 O(E√V)），内存 O(V+E)。

## 现在的行为

- `MaxMatching` 对每个左侧顶点只挑第一条还没被占用的边，不做增广：上面的样例给 `matching=1`（应 2）。
- `MinVertexCover` 返回 `ErrNoCover`，`--cover` 直接失败。
- `ParseEdges` 不校验顶点名，缺 `-` 只给一句笼统错误，没有边的下标。

## 计划新增的能力

```
--edges 边表    逗号分隔的 左-右
--cover         额外输出最小顶点覆盖
```

## 输出

```
matching=2
pairs=l1-r2,l2-r1
cover=l1,r1
```

## 目录

```
match.go              解析、匹配与顶点覆盖
cmd/bipartite-match   命令行入口
match_test.go         go test 用例
```
