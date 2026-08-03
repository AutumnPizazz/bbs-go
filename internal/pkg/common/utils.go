package common

import (
	"bbs-go/internal/models/constants"
	"bbs-go/internal/pkg/html"
	"bbs-go/internal/pkg/markdown"
	"bbs-go/internal/pkg/text"
)

// GetSummary 根据内容类型截取摘要
func GetSummary(contentType constants.ContentType, content string) (summary string) {
	if contentType == constants.ContentTypeMarkdown {
		summary = markdown.GetSummary(content, constants.SummaryLen)
	} else if contentType == constants.ContentTypeHtml {
		summary = html.GetSummary(content, constants.SummaryLen)
	} else {
		summary = text.GetSummary(content, constants.SummaryLen)
	}
	return
}

func Distinct[T any](input []T, getKey func(T) any) (output []T) {
	tempMap := map[any]byte{}
	for _, item := range input {
		l := len(tempMap)
		tempMap[getKey(item)] = 0
		if len(tempMap) != l { // 数量发生变化，说明不存在
			output = append(output, item)
		}
	}
	return
}
