package json_diff

import (
	"github.com/520MianXiangDuiXiang520/json-diff/decode"
)

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func longestCommonSubsequence(first, second []*decode.JsonNode) []*decode.JsonNode {
	line := len(first) + 1
	column := len(second) + 1
	if line == 1 || column == 1 {
		return make([]*decode.JsonNode, 0)
	}
	dp := make([][]int, line)
	for i := 0; i < line; i++ {
		dp[i] = make([]int, column)
	}
	for i := 1; i < line; i++ {
		for j := 1; j < column; j++ {
			if first[i-1].Equal(second[j-1]) {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			}
		}
	}
	// printDP(dp)
	start, end := len(first), len(second)
	cur := dp[start][end] - 1
	res := make([]*decode.JsonNode, cur+1)
	for cur >= 0 {
		if end >= 0 && start >= 0 &&
			dp[start][end] == dp[start][end-1] &&
			dp[start][end] == dp[start-1][end] {
			start--
			end--
		} else if end >= 0 && dp[start][end] == dp[start][end-1] {
			end--
		} else if start >= 0 && dp[start][end] == dp[start-1][end] {
			start--
		} else {
			res[cur] = first[start-1]
			cur--
			start--
			end--
		}
	}

	return res
}
