package render

import (
	"regexp"

	"bbs-go/internal/models/resp"
	"bbs-go/internal/pkg/search"
)

var htmlTagRe = regexp.MustCompile(`<[^>]*>`)

func stripHtml(s string) string {
	return htmlTagRe.ReplaceAllString(s, "")
}

func BuildSearchUsers(docs []search.UserDocument) []resp.SearchUserResponse {
	var items []resp.SearchUserResponse
	for _, doc := range docs {
		items = append(items, BuildSearchUser(doc))
	}
	return items
}

func BuildSearchUser(doc search.UserDocument) resp.SearchUserResponse {
	user := BuildUserInfoDefaultIfNull(doc.Id)
	// Bleve fragments may contain <mark> tags; strip them.
	nickname := stripHtml(doc.Nickname)
	desc := stripHtml(doc.Description)
	if nickname != "" {
		user.Nickname = nickname
	}
	if desc != "" {
		user.Description = desc
	}
	return resp.SearchUserResponse{
		User:        user,
		Nickname:    nickname,
		Username:    stripHtml(doc.Username),
		Description: desc,
		CreateTime:  doc.CreateTime,
	}
}
