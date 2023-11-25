package main

import (
	"database/sql"
	"errors"
	"net/http"
	"sort"
	"strconv"

	"github.com/jmoiron/sqlx"

	"github.com/labstack/echo/v4"
)

type UserStatsModel struct {
	UserID        int64 `db:"user_id"`
	ReactionCount int64 `db:"reaction_count"`
	TipCount      int64 `db:"tip_count"`
	CommentCount  int64 `db:"comment_count"`
	ViewerCount   int64 `db:"viewer_count"`
}

type LivestreamStatistics struct {
	Rank           int64 `json:"rank"`
	ViewersCount   int64 `json:"viewers_count"`
	TotalReactions int64 `json:"total_reactions"`
	TotalReports   int64 `json:"total_reports"`
	MaxTip         int64 `json:"max_tip"`
}

type LivestreamRankingEntry struct {
	LivestreamID int64
	Score        int64
}
type LivestreamRanking []LivestreamRankingEntry

func (r LivestreamRanking) Len() int      { return len(r) }
func (r LivestreamRanking) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r LivestreamRanking) Less(i, j int) bool {
	if r[i].Score == r[j].Score {
		return r[i].LivestreamID < r[j].LivestreamID
	} else {
		return r[i].Score < r[j].Score
	}
}

type UserStatistics struct {
	Rank              int64  `json:"rank"`
	ViewersCount      int64  `json:"viewers_count"`
	TotalReactions    int64  `json:"total_reactions"`
	TotalLivecomments int64  `json:"total_livecomments"`
	TotalTip          int64  `json:"total_tip"`
	FavoriteEmoji     string `json:"favorite_emoji"`
}

type UserRankingEntry struct {
	Username string
	Score    int64
}
type UserRanking []UserRankingEntry

func (r UserRanking) Len() int      { return len(r) }
func (r UserRanking) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r UserRanking) Less(i, j int) bool {
	if r[i].Score == r[j].Score {
		return r[i].Username < r[j].Username
	} else {
		return r[i].Score < r[j].Score
	}
}

func getUserStatisticsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	if err := verifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

	username := c.Param("username")
	// ユーザごとに、紐づく配信について、累計リアクション数、累計ライブコメント数、累計売上金額を算出
	// また、現在の合計視聴者数もだす

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to begin transaction: "+err.Error())
	}
	defer tx.Rollback()

	var user UserModel
	if err := tx.GetContext(ctx, &user, "SELECT * FROM users WHERE name = ?", username); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusBadRequest, "not found user that has the given username")
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get user: "+err.Error())
		}
	}

	// ランク算出
	var users []*UserModel
	if err := tx.SelectContext(ctx, &users, "SELECT * FROM users"); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get users: "+err.Error())
	}

	var ranking UserRanking
	for _, user := range users {
		var userStatsModel UserStatsModel
		foundUserStats := true
		if err := tx.GetContext(ctx, &userStatsModel, "SELECT * FROM user_stats WHERE user_id = ?", user.ID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				foundUserStats = false
			} else {
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to get user_stats: "+err.Error())
			}
		}
		var reactions int64
		var tips int64
		if foundUserStats {
			reactions = userStatsModel.ReactionCount
			tips = userStatsModel.TipCount
		}
		score := reactions + tips
		ranking = append(ranking, UserRankingEntry{
			Username: user.Name,
			Score:    score,
		})
	}
	sort.Sort(ranking)

	var rank int64 = 1
	for i := len(ranking) - 1; i >= 0; i-- {
		entry := ranking[i]
		if entry.Username == username {
			break
		}
		rank++
	}

	var userStatsModel UserStatsModel
	foundUserStats := true
	if err := tx.GetContext(ctx, &userStatsModel, "SELECT * FROM user_stats WHERE user_id = ?", user.ID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			foundUserStats = false
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get user_stats: "+err.Error())
		}
	}

	// リアクション数
	// ライブコメント数、チップ合計
	var totalReactions int64
	var totalLivecomments int64
	var totalTip int64
	var viewersCount int64
	if foundUserStats {
		totalReactions = userStatsModel.ReactionCount
		totalLivecomments = userStatsModel.CommentCount
		totalTip = userStatsModel.TipCount
		viewersCount = userStatsModel.ViewerCount
	}

	// お気に入り絵文字
	var favoriteEmoji string
	query := `
	SELECT r.emoji_name
	FROM users u
	INNER JOIN livestreams l ON l.user_id = u.id
	INNER JOIN reactions r ON r.livestream_id = l.id
	WHERE u.name = ?
	GROUP BY emoji_name
	ORDER BY COUNT(*) DESC, emoji_name DESC
	LIMIT 1
	`
	if err := tx.GetContext(ctx, &favoriteEmoji, query, username); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to find favorite emoji: "+err.Error())
	}

	stats := UserStatistics{
		Rank:              rank,
		ViewersCount:      viewersCount,
		TotalReactions:    totalReactions,
		TotalLivecomments: totalLivecomments,
		TotalTip:          totalTip,
		FavoriteEmoji:     favoriteEmoji,
	}
	return c.JSON(http.StatusOK, stats)
}

func getLivestreamStatisticsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	if err := verifyUserSession(c); err != nil {
		return err
	}

	id, err := strconv.Atoi(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "livestream_id in path must be integer")
	}
	livestreamID := int64(id)

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to begin transaction: "+err.Error())
	}
	defer tx.Rollback()

	var livestream LivestreamModel
	if err := tx.GetContext(ctx, &livestream, "SELECT * FROM livestreams WHERE id = ?", livestreamID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusBadRequest, "cannot get stats of not found livestream")
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get livestream: "+err.Error())
		}
	}

	var livestreams []*LivestreamModel
	if err := tx.SelectContext(ctx, &livestreams, "SELECT * FROM livestreams"); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get livestreams: "+err.Error())
	}

	// ランク算出

	// -- リアクション数計算
	livestreamIds := make([]int64, len(livestreams))
	for i := range livestreams {
		livestreamIds[i] = livestreams[i].ID
	}

	sqls := "SELECT * FROM reactions WHERE livestream_id IN (?)"
	sqls, params, err := sqlx.In(sqls, livestreamIds)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create sql: "+err.Error())
	}

	var reactionModels []*ReactionModel
	if err := tx.SelectContext(ctx, &reactionModels, sqls, params...); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get reactions: "+err.Error())
	}

	reactionsMap := make(map[int64]int64)
	for _, reaction := range reactionModels {
		reactionsMap[reaction.LivestreamID]++
	}

	// -- totalTips数
	sqls = "SELECT * FROM livecomments WHERE livestream_id IN (?)"
	sqls, params, err = sqlx.In(sqls, livestreamIds)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create sql: "+err.Error())
	}

	var livecommentModels []*LivecommentModel
	if err := tx.SelectContext(ctx, &livecommentModels, sqls, params...); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get reactions: "+err.Error())
	}

	// -- 最大チップ額
	maxTip := int64(0)
	totalTipsMap := make(map[int64]int64)
	for _, livecomment := range livecommentModels {
		totalTipsMap[livecomment.LivestreamID] += livecomment.Tip

		// livecommentのLivestreamIDがlivestream.IDと一致する場合、最大値を更新
		if livecomment.LivestreamID == livestream.ID {
			if livecomment.Tip > maxTip {
				maxTip = livecomment.Tip
			}
		}
	}

	var ranking LivestreamRanking
	for _, livestream := range livestreams {
		var reactions int64
		reactions = reactionsMap[livestream.ID]

		var totalTips int64
		totalTips = totalTipsMap[livestream.ID]

		score := reactions + totalTips
		ranking = append(ranking, LivestreamRankingEntry{
			LivestreamID: livestream.ID,
			Score:        score,
		})
	}
	sort.Sort(ranking)

	var rank int64 = 1
	for i := len(ranking) - 1; i >= 0; i-- {
		entry := ranking[i]
		if entry.LivestreamID == livestreamID {
			break
		}
		rank++
	}

	// 視聴者数算出
	var viewersCount int64
	if err := tx.GetContext(ctx, &viewersCount, `SELECT COUNT(*) FROM livestreams l INNER JOIN livestream_viewers_history h ON h.livestream_id = l.id WHERE l.id = ?`, livestreamID); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to count livestream viewers: "+err.Error())
	}

	// リアクション数
	var totalReactions int64
	totalReactions = reactionsMap[livestreamID]

	// スパム報告数
	var totalReports int64
	if err := tx.GetContext(ctx, &totalReports, `SELECT COUNT(*) FROM livestreams l INNER JOIN livecomment_reports r ON r.livestream_id = l.id WHERE l.id = ?`, livestreamID); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to count total spam reports: "+err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to commit: "+err.Error())
	}

	return c.JSON(http.StatusOK, LivestreamStatistics{
		Rank:           rank,
		ViewersCount:   viewersCount,
		MaxTip:         maxTip,
		TotalReactions: totalReactions,
		TotalReports:   totalReports,
	})
}
