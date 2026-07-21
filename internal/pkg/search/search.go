package search

import (
	"bbs-go/internal/cache"
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/pkg/config"
	html2 "bbs-go/internal/pkg/html"
	"bbs-go/internal/pkg/markdown"
	"bbs-go/internal/pkg/text"
	"bbs-go/internal/repositories"
	"html"
	"log/slog"
	"math"
	"time"

	"github.com/blevesearch/bleve/v2"
	blevequery "github.com/blevesearch/bleve/v2/search/query"
	"github.com/mitchellh/mapstructure"
	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/sqls"
)

var index bleve.Index

func Init() {
	var err error
	indexPath := config.Instance.Search.IndexPath
	if index, err = bleve.Open(indexPath); err != nil {
		if err == bleve.ErrorIndexPathDoesNotExist {
			index = newIndex(indexPath)
		} else {
			slog.Error(err.Error())
		}
	}
}

func NewTopicDoc(topic *models.Topic) *TopicDocument {
	if topic == nil {
		return nil
	}
	doc := &TopicDocument{
		Type:       EntityTypeTopic,
		Id:         topic.Id,
		CategoryId: topic.CategoryId,
		UserId:     topic.UserId,
		Format:     string(topic.Format),
		Title:      html.EscapeString(topic.Title),
		Summary:    html.EscapeString(topic.Summary),
		Status:     topic.Status,
		Recommend:  topic.Recommend,
		CreateTime: topic.CreateTime,
	}

	content := topic.Content
	if topic.ContentType == constants.ContentTypeMarkdown {
		content = markdown.ToHTML(content)
	}
	content = html.EscapeString(html2.GetHtmlText(content))
	doc.Content = content
	if strs.IsBlank(doc.Summary) {
		doc.Summary = text.GetSummary(content, constants.SummaryLen)
	}

	if user := cache.UserCache.Get(topic.UserId); user != nil {
		doc.Nickname = html.EscapeString(user.Nickname)
	}
	for _, tag := range getTopicTags(topic.Id) {
		doc.Tags = append(doc.Tags, tag.Name)
	}
	return doc
}

func NewUserDoc(user *models.User) *UserDocument {
	if user == nil {
		return nil
	}
	return &UserDocument{
		Type:         EntityTypeUser,
		Id:           user.Id,
		Username:     html.EscapeString(user.Username.String),
		Nickname:     html.EscapeString(user.Nickname),
		Avatar:       user.Avatar,
		Description:  html.EscapeString(user.Description),
		Status:       user.Status,
		TopicCount:   user.TopicCount,
		CommentCount: user.CommentCount,
		FansCount:    user.FansCount,
		FollowCount:  user.FollowCount,
		CreateTime:   user.CreateTime,
	}
}

func getTopicTags(topicId int64) []models.Tag {
	topicTags := repositories.TopicTagRepository.Find(sqls.DB(), sqls.NewCnd().Where("topic_id = ?", topicId))
	var tagIds []int64
	for _, topicTag := range topicTags {
		tagIds = append(tagIds, topicTag.TagId)
	}
	return cache.TagCache.GetList(tagIds)
}

func UpdateTopicIndexAsync(topic *models.Topic) {
	go UpdateTopicIndex(topic)
}

func UpdateTopicIndex(topic *models.Topic) {
	doc := NewTopicDoc(topic)
	if doc == nil || index == nil {
		return
	}
	if err := index.Index(searchDocID(EntityTypeTopic, topic.Id), doc); err != nil {
		slog.Error(err.Error())
	} else {
		slog.Info("add topic search index", slog.Any("id", topic.Id))
	}
}

func DeleteTopicIndex(id int64) error {
	if index == nil {
		return nil
	}
	return index.Delete(searchDocID(EntityTypeTopic, id))
}

func UpdateUserIndex(user *models.User) {
	doc := NewUserDoc(user)
	if doc == nil || index == nil {
		return
	}
	if err := index.Index(searchDocID(EntityTypeUser, user.Id), doc); err != nil {
		slog.Error(err.Error())
	} else {
		slog.Info("add user search index", slog.Any("id", user.Id))
	}
}

func DeleteUserIndex(id int64) error {
	if index == nil {
		return nil
	}
	return index.Delete(searchDocID(EntityTypeUser, id))
}

func SearchTopic(keyword string, categoryId int64, categoryIds []int64, timeRange int, format string, page, limit int) (docs []TopicDocument, paging *sqls.Paging, err error) {
	paging = &sqls.Paging{Page: page, Limit: limit}
	query := bleve.NewBooleanQuery()
	query.AddMust(bleve.NewMatchAllQuery())
	query.AddMust(typeQuery(EntityTypeTopic))

	if strs.IsNotBlank(format) {
		formatQuery := bleve.NewTermQuery(format)
		formatQuery.SetField("format")
		query.AddMust(formatQuery)
	}
	if strs.IsNotBlank(keyword) {
		query.AddMust(keywordQuery(keyword, []string{"title", "summary", "content", "tags", "nickname"}))
	}
	if categoryId != 0 {
		if categoryId == -1 {
			boolFieldQuery := bleve.NewBoolFieldQuery(true)
			boolFieldQuery.SetField("recommend")
			query.AddMust(boolFieldQuery)
		} else if categoryQuery := buildCategoryQuery(categoryId, categoryIds); categoryQuery != nil {
			query.AddMust(categoryQuery)
		}
	}
	addTimeRangeQuery(query, timeRange)

	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.From = paging.Offset()
	searchRequest.Size = paging.Limit
	searchRequest.Fields = []string{"*"}
	searchRequest.Highlight = bleve.NewHighlightWithStyle("html")
	searchRequest.Highlight.AddField("title")
	searchRequest.Highlight.AddField("summary")
	searchRequest.Highlight.AddField("content")

	result, err := index.Search(searchRequest)
	if err != nil {
		return nil, paging, err
	}
	for _, hit := range result.Hits {
		storedDoc := hitFields(hit.Fields, hit.Fragments)
		normalizeTags(storedDoc)
		var doc TopicDocument
		if err := mapstructure.Decode(storedDoc, &doc); err != nil {
			slog.Error(err.Error())
			continue
		}
		docs = append(docs, doc)
	}
	return docs, paging, nil
}

func SearchUser(keyword string, page, limit int) (docs []UserDocument, paging *sqls.Paging, err error) {
	paging = &sqls.Paging{Page: page, Limit: limit}
	query := bleve.NewBooleanQuery()
	query.AddMust(bleve.NewMatchAllQuery())
	query.AddMust(typeQuery(EntityTypeUser))
	if strs.IsNotBlank(keyword) {
		query.AddMust(keywordQuery(keyword, []string{"username", "nickname", "description"}))
	}
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.From = paging.Offset()
	searchRequest.Size = paging.Limit
	searchRequest.Fields = []string{"*"}
	searchRequest.Highlight = bleve.NewHighlightWithStyle("html")
	searchRequest.Highlight.AddField("nickname")
	searchRequest.Highlight.AddField("username")
	searchRequest.Highlight.AddField("description")

	result, err := index.Search(searchRequest)
	if err != nil {
		return nil, paging, err
	}
	for _, hit := range result.Hits {
		var doc UserDocument
		if err := mapstructure.Decode(hitFields(hit.Fields, hit.Fragments), &doc); err != nil {
			slog.Error(err.Error())
			continue
		}
		docs = append(docs, doc)
	}
	return docs, paging, nil
}

func SearchAll(keyword string, limit int) (AllResult, error) {
	if limit <= 0 {
		limit = 5
	}
	topics, _, err := SearchTopic(keyword, 0, nil, 0, "", 1, limit)
	if err != nil {
		return AllResult{}, err
	}
	users, _, err := SearchUser(keyword, 1, limit)
	if err != nil {
		return AllResult{}, err
	}
	return AllResult{Topics: topics, Users: users}, nil
}

func buildCategoryQuery(categoryId int64, categoryIds []int64) blevequery.Query {
	if len(categoryIds) == 0 {
		return buildExactCategoryQuery(categoryId)
	}
	if len(categoryIds) == 1 {
		return buildExactCategoryQuery(categoryIds[0])
	}
	queries := make([]blevequery.Query, 0, len(categoryIds))
	for _, id := range categoryIds {
		queries = append(queries, buildExactCategoryQuery(id))
	}
	return bleve.NewDisjunctionQuery(queries...)
}

func buildExactCategoryQuery(categoryId int64) blevequery.Query {
	f := float64(categoryId)
	b := true
	query := bleve.NewNumericRangeInclusiveQuery(&f, &f, &b, &b)
	query.SetField("categoryId")
	return query
}

func typeQuery(entityType string) blevequery.Query {
	query := bleve.NewTermQuery(entityType)
	query.SetField("type")
	return query
}

func keywordQuery(keyword string, fields []string) blevequery.Query {
	queries := make([]blevequery.Query, 0, len(fields))
	for _, field := range fields {
		query := bleve.NewMatchQuery(keyword)
		query.SetField(field)
		queries = append(queries, query)
	}
	return bleve.NewDisjunctionQuery(queries...)
}

func addTimeRangeQuery(query *blevequery.BooleanQuery, timeRange int) {
	if timeRange == 0 {
		return
	}
	var beginTime int64
	switch timeRange {
	case 1:
		beginTime = dates.Timestamp(time.Now().Add(-24 * time.Hour))
	case 2:
		beginTime = dates.Timestamp(time.Now().Add(-7 * 24 * time.Hour))
	case 3:
		beginTime = dates.Timestamp(time.Now().AddDate(0, -1, 0))
	case 4:
		beginTime = dates.Timestamp(time.Now().AddDate(-1, 0, 0))
	}
	if beginTime == 0 {
		return
	}
	min := float64(beginTime)
	max := float64(math.MaxInt64)
	createTimeQuery := bleve.NewNumericRangeQuery(&min, &max)
	createTimeQuery.SetField("createTime")
	query.AddMust(createTimeQuery)
}

func hitFields(fields map[string]interface{}, fragments map[string][]string) map[string]interface{} {
	storedDoc := make(map[string]interface{})
	for key, field := range fields {
		storedDoc[key] = field
	}
	for field, values := range fragments {
		if len(values) > 0 {
			storedDoc[field] = values[0]
		}
	}
	return storedDoc
}

func normalizeTags(storedDoc map[string]interface{}) {
	tagField, ok := storedDoc["tags"]
	if !ok {
		return
	}
	switch value := tagField.(type) {
	case string:
		storedDoc["tags"] = []string{value}
	case []interface{}:
		tags := make([]string, 0, len(value))
		for _, tag := range value {
			if name, ok := tag.(string); ok {
				tags = append(tags, name)
			}
		}
		storedDoc["tags"] = tags
	}
}
