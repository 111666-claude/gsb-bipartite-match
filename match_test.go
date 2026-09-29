package bipartitematch

import (
	"errors"
	"fmt"
	"math/bits"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestParseEdges(t *testing.T) {
	edges, err := ParseEdges("l1-r1, l1-r2")
	if err != nil {
		t.Fatalf("ParseEdges 出错：%v", err)
	}
	if len(edges) != 2 || edges[0] != (Edge{Left: "l1", Right: "r1"}) || edges[1] != (Edge{Left: "l1", Right: "r2"}) {
		t.Fatalf("解析结果不对：%+v", edges)
	}
}

func TestParseEdgesRejectsMissingDash(t *testing.T) {
	if _, err := ParseEdges("l1r1"); err == nil {
		t.Fatal("缺 - 应该报错")
	}
}

func TestSingleEdge(t *testing.T) {
	edges, _ := ParseEdges("l1-r1")
	pairs := MaxMatching(edges)
	if len(pairs) != 1 || pairs[0] != (Edge{Left: "l1", Right: "r1"}) {
		t.Fatalf("单条边应给一个配对：%+v", pairs)
	}
}

func TestDisjointEdges(t *testing.T) {
	edges, _ := ParseEdges("l1-r1,l2-r2")
	if pairs := MaxMatching(edges); len(pairs) != 2 {
		t.Fatalf("两条不相交的边应给两个配对：%+v", pairs)
	}
}

func TestMatchingIsValid(t *testing.T) {
	edges, _ := ParseEdges("l1-r1,l1-r2,l2-r1,l2-r2")
	lefts := map[string]bool{}
	rights := map[string]bool{}
	for _, pair := range MaxMatching(edges) {
		if lefts[pair.Left] || rights[pair.Right] {
			t.Fatalf("配对里出现了重复顶点：%+v", pair)
		}
		lefts[pair.Left] = true
		rights[pair.Right] = true
	}
}

func TestMatchingSizeIsAtMostSide(t *testing.T) {
	edges, _ := ParseEdges("l1-r1,l1-r2,l1-r3")
	if pairs := MaxMatching(edges); len(pairs) > 1 {
		t.Fatalf("左侧只有一个顶点，匹配不能超过 1：%+v", pairs)
	}
}

func TestParseEdgesSkipsBlankItems(t *testing.T) {
	edges, err := ParseEdges("l1-r1,")
	if err != nil {
		t.Fatalf("ParseEdges 出错：%v", err)
	}
	if len(edges) != 1 {
		t.Fatalf("空项应该被跳过：%+v", edges)
	}
}

func TestAugmentingPathBeatsGreedy(t *testing.T) {
	edges, err := ParseEdges("l1-r1,l1-r2,l2-r1")
	if err != nil {
		t.Fatalf("ParseEdges 出错：%v", err)
	}
	pairs := MaxMatching(edges)
	want := []Edge{{Left: "l1", Right: "r2"}, {Left: "l2", Right: "r1"}}
	if len(pairs) != len(want) {
		t.Fatalf("应通过增广路得到 matching=2，实际 %+v", pairs)
	}
	for i := range want {
		if pairs[i] != want[i] {
			t.Fatalf("第 %d 个配对应为 %+v，实际 %+v", i, want[i], pairs[i])
		}
	}
}

func TestMinVertexCoverSample(t *testing.T) {
	edges, _ := ParseEdges("l1-r1,l1-r2,l2-r1")
	cover := MinVertexCover(edges)
	if len(cover) != 2 {
		t.Fatalf("最小顶点覆盖大小应为 2，实际 %+v", cover)
	}
	assertCoverValid(t, edges, cover)
}

func TestMinVertexCoverEmpty(t *testing.T) {
	if cover := MinVertexCover(nil); len(cover) != 0 {
		t.Fatalf("空图的覆盖应为空，实际 %+v", cover)
	}
}

func TestParseEdgesReportsIndex(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"缺 - 在第 1 条", "l1r1", "第 1 条"},
		{"缺 - 在第 2 条", "l1-r1,l2r2", "第 2 条"},
		{"空左顶点", "-r1", "第 1 条"},
		{"空右顶点在第 2 条", "l1-r1,l2-", "第 2 条"},
		{"重复边", "l1-r1,l2-r2,l1-r1", "第 3 条"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseEdges(tc.input)
			if err == nil {
				t.Fatalf("%q 应该报错", tc.input)
			}
			if !errors.Is(err, ErrFormat) {
				t.Fatalf("错误应包装 ErrFormat：%v", err)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("错误应包含 %q：%v", tc.want, err)
			}
		})
	}
}

func TestDeterministic(t *testing.T) {
	edges, _ := ParseEdges("b-y,a-x,a-y,b-x,c-x")
	firstPairs := MaxMatching(edges)
	firstCover := MinVertexCover(edges)
	for i := 0; i < 10; i++ {
		if !reflect.DeepEqual(MaxMatching(edges), firstPairs) {
			t.Fatal("同一输入重复运行，pairs 不一致")
		}
		if !reflect.DeepEqual(MinVertexCover(edges), firstCover) {
			t.Fatal("同一输入重复运行，cover 不一致")
		}
	}
}

// assertCoverValid 校验 cover 覆盖每一条边。
func assertCoverValid(t *testing.T, edges []Edge, cover []string) {
	t.Helper()
	inCover := map[string]bool{}
	for _, name := range cover {
		inCover[name] = true
	}
	for _, edge := range edges {
		if !inCover[edge.Left] && !inCover[edge.Right] {
			t.Fatalf("cover %+v 没覆盖边 %+v", cover, edge)
		}
	}
}

// bruteForceMatching 穷举所有匹配取最大（仅用于小规模对照）。
func bruteForceMatching(edges []Edge) int {
	adj := map[string][]string{}
	lefts := []string{}
	seenLeft := map[string]bool{}
	for _, edge := range edges {
		adj[edge.Left] = append(adj[edge.Left], edge.Right)
		if !seenLeft[edge.Left] {
			seenLeft[edge.Left] = true
			lefts = append(lefts, edge.Left)
		}
	}
	usedRight := map[string]bool{}
	var search func(i int) int
	search = func(i int) int {
		if i == len(lefts) {
			return 0
		}
		best := search(i + 1)
		for _, right := range adj[lefts[i]] {
			if usedRight[right] {
				continue
			}
			usedRight[right] = true
			if size := 1 + search(i+1); size > best {
				best = size
			}
			usedRight[right] = false
		}
		return best
	}
	return search(0)
}

// bruteForceCover 枚举顶点子集取最小覆盖（仅用于小规模对照）。
func bruteForceCover(edges []Edge) int {
	names := []string{}
	seen := map[string]bool{}
	for _, edge := range edges {
		for _, name := range []string{edge.Left, edge.Right} {
			if !seen[name] {
				seen[name] = true
				names = append(names, name)
			}
		}
	}
	index := map[string]int{}
	for i, name := range names {
		index[name] = i
	}
	best := len(names) + 1
	for mask := 0; mask < 1<<len(names); mask++ {
		size := bits.OnesCount(uint(mask))
		if size >= best {
			continue
		}
		covered := true
		for _, edge := range edges {
			if mask&(1<<index[edge.Left]) == 0 && mask&(1<<index[edge.Right]) == 0 {
				covered = false
				break
			}
		}
		if covered {
			best = size
		}
	}
	return best
}

func TestAgainstBruteForce(t *testing.T) {
	rng := rand.New(rand.NewSource(20260929))
	mismatch := 0
	for trial := 0; trial < 300; trial++ {
		nL := 1 + rng.Intn(6)
		nR := 1 + rng.Intn(6)
		edges := []Edge{}
		for l := 0; l < nL; l++ {
			for r := 0; r < nR; r++ {
				if rng.Float64() < 0.4 {
					edges = append(edges, Edge{
						Left:  fmt.Sprintf("l%d", l),
						Right: fmt.Sprintf("r%d", r),
					})
				}
			}
		}
		pairs := MaxMatching(edges)
		cover := MinVertexCover(edges)

		// 不变量：配对左右两两不同，大小不超过两侧顶点数。
		usedL, usedR := map[string]bool{}, map[string]bool{}
		for _, pair := range pairs {
			if usedL[pair.Left] || usedR[pair.Right] {
				t.Fatalf("trial %d：配对出现重复顶点 %+v", trial, pair)
			}
			usedL[pair.Left] = true
			usedR[pair.Right] = true
		}
		if len(pairs) > nL || len(pairs) > nR {
			t.Fatalf("trial %d：匹配大小 %d 超过两侧顶点数 %d/%d", trial, len(pairs), nL, nR)
		}

		// 不变量：cover 覆盖所有边且大小等于最大匹配。
		assertCoverValid(t, edges, cover)
		if len(cover) != len(pairs) {
			t.Fatalf("trial %d：|cover|=%d != matching=%d", trial, len(cover), len(pairs))
		}

		// 对照穷举：差异必须为 0。
		if got, want := len(pairs), bruteForceMatching(edges); got != want {
			mismatch++
			t.Errorf("trial %d：matching=%d，穷举=%d，edges=%+v", trial, got, want, edges)
		}
		if got, want := len(cover), bruteForceCover(edges); got != want {
			mismatch++
			t.Errorf("trial %d：cover=%d，穷举=%d，edges=%+v", trial, got, want, edges)
		}
	}
	if mismatch != 0 {
		t.Fatalf("与穷举的差异必须为 0，实际 %d", mismatch)
	}
}

func TestLargeGraph(t *testing.T) {
	if testing.Short() {
		t.Skip("短模式跳过规模测试")
	}
	const nL, nR, nE = 10000, 10000, 100000
	rng := rand.New(rand.NewSource(42))
	edges := make([]Edge, 0, nE)
	seen := map[[2]int]bool{}
	for len(edges) < nE {
		l, r := rng.Intn(nL), rng.Intn(nR)
		if seen[[2]int{l, r}] {
			continue
		}
		seen[[2]int{l, r}] = true
		edges = append(edges, Edge{
			Left:  fmt.Sprintf("l%d", l),
			Right: fmt.Sprintf("r%d", r),
		})
	}
	start := time.Now()
	pairs := MaxMatching(edges)
	cover := MinVertexCover(edges)
	elapsed := time.Since(start)
	if len(pairs) > nL || len(pairs) > nR {
		t.Fatalf("匹配大小 %d 超过两侧顶点数", len(pairs))
	}
	assertCoverValid(t, edges, cover)
	if len(cover) != len(pairs) {
		t.Fatalf("|cover|=%d != matching=%d", len(cover), len(pairs))
	}
	if elapsed > 3*time.Second {
		t.Fatalf("1 万点 10 万边应在 3 秒内出结果，实际 %v", elapsed)
	}
	t.Logf("matching=%d cover=%d 耗时 %v", len(pairs), len(cover), elapsed)
}
